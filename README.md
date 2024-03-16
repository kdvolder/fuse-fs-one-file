Fuse FS One File
================

Create a 'fuse' filesystem that behaves like it contains one single giant file.
The file is initially filled with zeros, but these zeros are not stored.
Only parts of the file that actually are written to and contain non-0 data
will be actually stored.

The idea is that such a file can be mounted as a loop device and formatted with
a real filesystem such as ext4 or btrfs. 

We want to ultimately store the data for this giant file onto a cloud-storage system,
in particular 'box'. The problem with using box on its own or via rclone, is that
it is extremely slow when working with many small files. It performs much better
when data is wread and written in large quantities.

Thus we will used it as a backing store to represent large blocks of data which 
contain sections of our 'biG disk image' file.

We can also cache the data locally for faster read and write.

Plan
====

Stage 1: local block file storage.

- Start with readng article: Create your own 'fuse' filesystem in go: https://blog.gopheracademy.com/advent-2014/fuse-zipfs/
- Try to implement a fuse fs similar to the above but which:
   - presents a single large file called 'disk.img' to the user
   - uses a dsignated local directory as a backing store
   - stores data in the backing store directory as numbered 'block' files 'blk-00001', 'blk-00002' etc.
   - only the block files that have non-0 data are written on disk. 
   - the size of a 'block' is configurable on file-system creation
   - the total size of the file is configurable on fs creation
   - the backing-storage directory is configurable on fs creation.
   - validation: the backing store must be empty for fs creation.
   - a `onefs.json` metadata file is saved to the backing store with any needed infos to successfully mount the fs. 

Stage 2: create an abstraction that allows the implementation of different 'storage backends'

- ....

