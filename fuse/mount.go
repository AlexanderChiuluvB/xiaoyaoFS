package fuse

import (
	"fmt"
	"github.com/seaweedfs/fuse/fs"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"github.com/seaweedfs/fuse"
	"time"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/command"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/filer"

)

type MountOptions struct {
	filer                       *string
	filerMountRootPath          *string
	dir                         *string
	dirAutoCreate               *bool
	collection                  *string
	replication                 *string
	ttlSec                      *int
	chunkSizeLimitMB            *int
	cacheDir                    *string
	cacheSizeMB                 *int64
	dataCenter                  *string
	allowOthers                 *bool
	umaskString                 *string
	nonempty                    *bool
	outsideContainerClusterMode *bool
	uidMap                      *string
	gidMap                      *string
}

var (
	mountOptions       MountOptions
	mountCpuProfile    *string
	mountMemProfile    *string
	mountReadRetryTime *time.Duration
)

func init() {
	cmdMount.Run = runMount // break init cycle
	mountOptions.filer = cmdMount.Flag.String("filer", "localhost:8888", "weed filer location")
	mountOptions.filerMountRootPath = cmdMount.Flag.String("filer.path", "/", "mount this remote path from filer server")
	mountOptions.dir = cmdMount.Flag.String("dir", ".", "mount weed filer to this directory")
	mountOptions.dirAutoCreate = cmdMount.Flag.Bool("dirAutoCreate", false, "auto create the directory to mount to")
	mountOptions.collection = cmdMount.Flag.String("collection", "", "collection to create the files")
	mountOptions.replication = cmdMount.Flag.String("replication", "", "replication(e.g. 000, 001) to create to files. If empty, let filer decide.")
	mountOptions.ttlSec = cmdMount.Flag.Int("ttl", 0, "file ttl in seconds")
	mountOptions.chunkSizeLimitMB = cmdMount.Flag.Int("chunkSizeLimitMB", 16, "local write buffer size, also chunk large files")
	mountOptions.cacheDir = cmdMount.Flag.String("cacheDir", os.TempDir(), "local cache directory for file chunks and meta data")
	mountOptions.cacheSizeMB = cmdMount.Flag.Int64("cacheCapacityMB", 1000, "local file chunk cache capacity in MB (0 will disable cache)")
	mountOptions.dataCenter = cmdMount.Flag.String("dataCenter", "", "prefer to write to the data center")
	mountOptions.allowOthers = cmdMount.Flag.Bool("allowOthers", true, "allows other users to access the file system")
	mountOptions.umaskString = cmdMount.Flag.String("umask", "022", "octal umask, e.g., 022, 0111")
	mountOptions.nonempty = cmdMount.Flag.Bool("nonempty", false, "allows the mounting over a non-empty directory")
	mountOptions.outsideContainerClusterMode = cmdMount.Flag.Bool("outsideContainerClusterMode", false, "allows other users to access volume servers with publicUrl")
	mountOptions.uidMap = cmdMount.Flag.String("map.uid", "", "map local uid to uid on filer, comma-separated <local_uid>:<filer_uid>")
	mountOptions.gidMap = cmdMount.Flag.String("map.gid", "", "map local gid to gid on filer, comma-separated <local_gid>:<filer_gid>")

	mountCpuProfile = cmdMount.Flag.String("cpuprofile", "", "cpu profile output file")
	mountMemProfile = cmdMount.Flag.String("memprofile", "", "memory profile output file")
	mountReadRetryTime = cmdMount.Flag.Duration("readRetryTime", 6*time.Second, "maximum read retry wait time")
}

var cmdMount = &command.Command{
	UsageLine: "mount -filer=localhost:8888 -dir=/some/dir",
	Short:     "mount weed filer to a directory as file system in userspace(FUSE)",
	Long: `mount weed filer to userspace.

  Pre-requisites:
  1) have xiaoyaoFS master and volume servers running
  2) have a "xiaoyao filer" running
  These 2 requirements can be achieved with one command "weed server -filer=true"

  `,
}

func runMount(cmd *command.Command, args []string) bool {

	grace.SetupProfiling(*mountCpuProfile, *mountMemProfile)
	if *mountReadRetryTime < time.Second {
		*mountReadRetryTime = time.Second
	}
	filer.ReadWaitTime = *mountReadRetryTime

	umask, umaskErr := strconv.ParseUint(*mountOptions.umaskString, 8, 64)
	if umaskErr != nil {
		fmt.Printf("can not parse umask %s", *mountOptions.umaskString)
		return false
	}

	if len(args) > 0 {
		return false
	}

	return RunMount(&mountOptions, os.FileMode(umask))
}

func RunMount(option *MountOptions, umask os.FileMode) bool {

	filer := *option.filer
	// parse filer grpc address
	filerGrpcAddress, err := pb.ParseFilerGrpcAddress(filer)
	if err != nil {
		glog.V(0).Infof("ParseFilerGrpcAddress: %v", err)
		return true
	}

	// try to connect to filer, filerBucketsPath may be useful later
	grpcDialOption := security.LoadClientTLS(util.GetViper(), "grpc.client")
	var cipher bool
	err = pb.WithGrpcFilerClient(filerGrpcAddress, grpcDialOption, func(client filer_pb.SeaweedFilerClient) error {
		resp, err := client.GetFilerConfiguration(context.Background(), &filer_pb.GetFilerConfigurationRequest{})
		if err != nil {
			return fmt.Errorf("get filer grpc address %s configuration: %v", filerGrpcAddress, err)
		}
		cipher = resp.Cipher
		return nil
	})
	if err != nil {
		glog.Infof("failed to talk to filer %s: %v", filerGrpcAddress, err)
		return true
	}

	filerMountRootPath := *option.filerMountRootPath
	dir := util.ResolvePath(*option.dir)
	chunkSizeLimitMB := *mountOptions.chunkSizeLimitMB

	fmt.Printf("This is SeaweedFS version %s %s %s\n", util.Version(), runtime.GOOS, runtime.GOARCH)
	if dir == "" {
		fmt.Printf("Please specify the mount directory via \"-dir\"")
		return false
	}
	if chunkSizeLimitMB <= 0 {
		fmt.Printf("Please specify a reasonable buffer size.")
		return false
	}

	fuse.Unmount(dir)

	// detect mount folder mode
	if *option.dirAutoCreate {
		os.MkdirAll(dir, os.FileMode(0777)&^umask)
	}
	fileInfo, err := os.Stat(dir)

	uid, gid := uint32(0), uint32(0)
	mountMode := os.ModeDir | 0777
	if err == nil {
		mountMode = os.ModeDir | fileInfo.Mode()
		uid, gid = util.GetFileUidGid(fileInfo)
		fmt.Printf("mount point owner uid=%d gid=%d mode=%s\n", uid, gid, fileInfo.Mode())
	} else {
		fmt.Printf("can not stat %s\n", dir)
		return false
	}

	if uid == 0 {
		if u, err := user.Current(); err == nil {
			if parsedId, pe := strconv.ParseUint(u.Uid, 10, 32); pe == nil {
				uid = uint32(parsedId)
			}
			if parsedId, pe := strconv.ParseUint(u.Gid, 10, 32); pe == nil {
				gid = uint32(parsedId)
			}
			fmt.Printf("current uid=%d gid=%d\n", uid, gid)
		}
	}

	// mapping uid, gid
	uidGidMapper, err := meta_cache.NewUidGidMapper(*option.uidMap, *option.gidMap)
	if err != nil {
		fmt.Printf("failed to parse %s %s: %v\n", *option.uidMap, *option.gidMap, err)
		return false
	}

	mountName := path.Base(dir)

	options := []fuse.MountOption{
		fuse.VolumeName(mountName),
		fuse.FSName(filer + ":" + filerMountRootPath),
		fuse.Subtype("xiaoyao"),
		// fuse.NoAppleDouble(), // include .DS_Store, otherwise can not delete non-empty folders
		fuse.NoAppleXattr(),
		fuse.NoBrowse(),
		fuse.AutoXattr(),
		fuse.ExclCreate(),
		fuse.DaemonTimeout("3600"),
		fuse.AllowSUID(),
		fuse.DefaultPermissions(),
		fuse.MaxReadahead(1024 * 128),
		fuse.AsyncRead(),
		fuse.WritebackCache(),
	}

	// find mount point
	mountRoot := filerMountRootPath
	if mountRoot != "/" && strings.HasSuffix(mountRoot, "/") {
		mountRoot = mountRoot[0 : len(mountRoot)-1]
	}

	xiaoyaoFileSystem := filesys.NewXiaoyaoFileSystem(&filesys.Option{
		FilerGrpcAddress:            filerGrpcAddress,
		GrpcDialOption:              grpcDialOption,
		FilerMountRootPath:          mountRoot,
		Collection:                  *option.collection,
		Replication:                 *option.replication,
		TtlSec:                      int32(*option.ttlSec),
		ChunkSizeLimit:              int64(chunkSizeLimitMB) * 1024 * 1024,
		CacheDir:                    *option.cacheDir,
		CacheSizeMB:                 *option.cacheSizeMB,
		DataCenter:                  *option.dataCenter,
		EntryCacheTtl:               3 * time.Second,
		MountUid:                    uid,
		MountGid:                    gid,
		MountMode:                   mountMode,
		MountCtime:                  fileInfo.ModTime(),
		MountMtime:                  time.Now(),
		Umask:                       umask,
		OutsideContainerClusterMode: *mountOptions.outsideContainerClusterMode,
		Cipher:                      cipher,
		UidGidMapper:                uidGidMapper,
	})

	// mount
	c, err := fuse.Mount(dir, options...)
	defer fuse.Unmount(dir)

	err = fs.Serve(c, seaweedFileSystem)

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Printf("mount process: %v", err)
		return true
	}

	return true
}