# WARNING #

**This code is not being used anymore.  I have replaced everything with [btsync](http://www.bittorrent.com/sync) and some very simple rdiff-backup scripts in jenkins.**

# Overview #

Given a source and destination folder and a task, goodnewseveryone executes a command between two, possibly remotely shared, folders.

I use goodnewseveryone with rdiff-backup, unison and rsync to sync and backup my photos and other files between multiple computers automatically from a single server using [jenkins](http://jenkins-ci.org/).

Goodnewseveryone has been tested on ubuntu 12.04, but should be runnable on quite a few linux distributions.

# Installation #

  * Install http://golang.org/
  * Create a folder for goodnewseveryone anywhere, for example /home/walter/gne
  * cd /home/walter/gne
  * mkdir src
  * cd src
  * git clone https://code.google.com/p/goodnewseveryone
  * export GOPATH=/home/walter/gne
  * cd goodnewseveryone
  * go build .
  * Copy the goodnewseveryone binary somewhere into the root user's PATH

# Getting Started #

Goodnewseveryone expects a json file to be piped into its standard in, for example:

```
{
	"Src": {
		"Mount": "mount -o port=6553,guest -t cifs //{{.IPAddress}}/{{.Remote}} {{.Local}}",
		"Unmount": "umount -l {{.Local}}",
		"IPAddress": "localhost",
		"Username": "",
		"Password": "",
		"Remote": "testremote",
		"Local": "/media/testremote/"
	},
	"Dst": {
		"Mount": "",
		"Unmount": "",
		"IPAddress": "",
		"Username": "",
		"Password": "",
		"Remote": "",
		"Local": "./testlocal/"
	},
	"Task": "rsync -r {{.Src.Local}} {{.Dst.Local}}"
} | sudo goodnewseveryone
```

Goodnewseveryone uses [gotemplates](http://172.16.63.94:6060/pkg/text/template/) to describe the Mount, Unmount and Task parameters in a general way.  Goodnewseveryone requires sudo rights for mounting and unmounting.

Goodnewseveryone is a very small program which basically only executes the following logic:

![https://chart.googleapis.com/chart?cht=gv&chl=digraph{%22Unmount%20Src%22%20-%3E%20%22Unmount%20Dst%22;%20%22Unmount%20Dst%22%20-%3E%20%22Src.Local%20Exists%22;%20%22Src.Local%20Exists%22%20-%3E%20%22Create%20Src.Local%22%20[label=%22No%22];%20%22Src.Local%20Exists%22%20-%3E%20%22Dst.Local%20Exists%22%20[label=%22Yes%22];%20%22Create%20Src.Local%22%20-%3E%20%22Dst.Local%20Exists%22;%20%22Dst.Local%20Exists%22%20-%3E%20%22Create%20Dst.Local%22%20[label=%22No%22];%20%22Create%20Dst.Local%22%20-%3E%20%22Mount%20Src%22;%20%22Dst.Local%20Exists%22%20-%3E%20%22Mount%20Src%22%20[label=%22Yes%22];%20%22Mount%20Src%22%20-%3E%20%22Error%22%20[label=%22Error%22];%20%22Mount%20Src%22%20-%3E%20%22Mount%20Dst%22;%20%22Mount%20Dst%22%20-%3E%20%22Unmount%20Src%20(Error)%22%20[label=%22Error%22];%20%22Unmount%20Src%20(Error)%22%20-%3E%20%22Error%22;%20%22Mount%20Dst%22%20-%3E%20%22Execute%20Task%22;%20%22Execute%20Task%22%20-%3E%20%22Unmount%20Dst%20(Success)%22;%20%22Unmount%20Dst%20(Success)%22%20-%3E%20%22Umount%20Src%20(Success)%22;%20%22Umount%20Src%20(Success)%22%20-%3E%20%22Success%22;%20%22Execute%20Task%22%20-%3E%20%22Unmount%20Dst%20(Error)%22%20[label=%22Error%22];%20%22Unmount%20Dst%20(Error)%22%20-%3E%20%22Unmount%20Src%20(Error)%22;}&nonsense=something_that_ends_with.png](https://chart.googleapis.com/chart?cht=gv&chl=digraph{%22Unmount%20Src%22%20-%3E%20%22Unmount%20Dst%22;%20%22Unmount%20Dst%22%20-%3E%20%22Src.Local%20Exists%22;%20%22Src.Local%20Exists%22%20-%3E%20%22Create%20Src.Local%22%20[label=%22No%22];%20%22Src.Local%20Exists%22%20-%3E%20%22Dst.Local%20Exists%22%20[label=%22Yes%22];%20%22Create%20Src.Local%22%20-%3E%20%22Dst.Local%20Exists%22;%20%22Dst.Local%20Exists%22%20-%3E%20%22Create%20Dst.Local%22%20[label=%22No%22];%20%22Create%20Dst.Local%22%20-%3E%20%22Mount%20Src%22;%20%22Dst.Local%20Exists%22%20-%3E%20%22Mount%20Src%22%20[label=%22Yes%22];%20%22Mount%20Src%22%20-%3E%20%22Error%22%20[label=%22Error%22];%20%22Mount%20Src%22%20-%3E%20%22Mount%20Dst%22;%20%22Mount%20Dst%22%20-%3E%20%22Unmount%20Src%20(Error)%22%20[label=%22Error%22];%20%22Unmount%20Src%20(Error)%22%20-%3E%20%22Error%22;%20%22Mount%20Dst%22%20-%3E%20%22Execute%20Task%22;%20%22Execute%20Task%22%20-%3E%20%22Unmount%20Dst%20(Success)%22;%20%22Unmount%20Dst%20(Success)%22%20-%3E%20%22Umount%20Src%20(Success)%22;%20%22Umount%20Src%20(Success)%22%20-%3E%20%22Success%22;%20%22Execute%20Task%22%20-%3E%20%22Unmount%20Dst%20(Error)%22%20[label=%22Error%22];%20%22Unmount%20Dst%20(Error)%22%20-%3E%20%22Unmount%20Src%20(Error)%22;}&nonsense=something_that_ends_with.png)

# Examples #

## Samba ##

The mount and unmount parameters for samba can be set in many ways.
Here are some parameters that I found to be quite OS agnostic.
```
Mount: "mount -o username={{.Username}},password={{.Password}},nounix,noserverino,sec=ntlmssp -t cifs //{{.IPAddress}}/{{.Remote}} {{.Local}}"
Unmount: "umount -l {{.Local}}"
```
Although sometimes I prefer
```
Mount: "mount -o username={{.Username}},password={{.Password}},iocharset=utf8,mode=0777,file_mode=0777 -t cifs //{{.IPAddress}}/{{.Remote}} {{.Local}}"
Unmount : "umount -l {{.Local}}"
```

## FTP ##

Running a task with a shared ftp folder can be done by installing curlftpfs in ubuntu 12.04
```
sudo apt-get install curlftpfs
```
The mount and unmount parameters can then be set as below:
```
Mount: "curlftpfs {{.Username}}:{{.Password}}@{{.IPAddress}}/{{.Remote}} {{.Local}}"
UnMount: "fusermount -u {{.Local}}"
```

## USB ##

I like to plug in an extra Backup USB Drive into my server on occasion.
These parameters allow me to mount my NTFS formatted USB Drive and run a backup script and finally unmount the drive.
```
Mount: "mount -t ntfs-3g UUID=\"{{.Remote}}\" {{.Local}}"
Unmount: umount -l {{.Local}}"
```
Using this with jenkins I can simply plug in my USB Drive and walk away, or go to the jenkins web interface and tell the job to execute.
Here I have set the Remote parameter to be the USB Drive UUID, which I have obtained with the blkid command.

## Backup ##

I like to backup my documents with the following task

```
"Task": "rdiff-backup {{.Src.Local}} {{.Dst.Local}}"
```

## Mirror ##

I like to create a mirror of some backups on a USB Drive with this task:
```
"Task": "rsync -r --delete {{.Src.Local}} {{.Dst.Local}}"
```

## Move ##

This is the task I execute the most to move new files from my laptop to my server.
I really hated waiting for a copy to execute and would rather let my server do it.
This repeated frustration with OSX super slow samba implementation and syncing photos was the inspiration for this project.
```
"Task": "rsync -r --remove-source-files {{.Src.Local}} {{.Dst.Local}}"
```

## Sync ##

Finally I like to sync my photos between computers with this task:
```
"Task": "unison -fastcheck true -batch -dontchmod -perms 0 {{.Src.Local}} {{.Dst.Local}}"
```

# Usage with Jenkins #

[Jenkins](http://jenkins-ci.org/) is an extendable open source continuous integration server.
It is typically used for building code projects and running their tests, but basically it manages jobs.
These jobs are commands which are executed.  Jenkins keeps logs and can periodically run these jobs and much more.
All Goodnewseveryone needs is someone to execute it and this, in my case, is jenkins.

Installing jenkins is really easy, please see their website.
Since Goodnewseveryone needs sudo rights you will have give jenkins sudo rights.
The easiest way I found to do this was to add jenkins to the sudoers file.
Ok now jenkins is setup.

First I share the folders which I would like to run jobs between.
Then I use my router to assign static IPs to my laptops' mac addresses.

Next I create an "ishome" job for each laptop in my home.
I let these jobs run periodically every hour.
These "ishome" jobs use the ping command.
For example:
```
ping -c 1 192.168.1.1
```
The idea is that this job will fail when your laptop is not home and succeed when your laptop home.

Next I create a job which will run goodnewseveryone.
This job will only run when the corresponding ishome job succeeds.
This means that the backup, sync etc. for photos, code, etc. will only run when my laptop is on the network.

Now I get that there are a lot of parameters, but this is where I start using environment variables which jenkins also has and it allows me to create a quite compact command for each job.

```
{
	"Src": {
                $SMB,
                $WALTERPHOTOS,
                "Remote": "Photos",
		"Local": "/media/WalterPhotos/"
        },
        "Dst": {
                "Local": "/media/Backup/Photos/"
        },
        Task: $BACKUP,
} | $GNE

```

# Troubleshooting #

## 2017 ##

If you try to mount a Windows 7 computer using ubuntu you might get the following error
```
mount error(12): Cannot allocate memory
```
In this case run the [2017fix.reg](https://code.google.com/p/goodnewseveryone/source/browse/2017fix.reg) registry fix file on the windows computer and it should fix the problem.

For more information and the reason see http://alan.lamielle.net/2009/09/03/windows-7-nonpaged-pool-srv-error-2017