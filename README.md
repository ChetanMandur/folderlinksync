# Folder-Link-Sync (FLS) 

## A simple program to sync files between two directories
[!CAUTION]
*Currently in deep alpha-v1.0*

[!WARNING]
<ins>**While files are generally safe from corruption/deletion, many advance features may not work as intended**</ins>

***This project was left unmaintained for quite some time, but is now picking up steam again***

**The v1.0 release will aim to just cleanup and fix the current code and functionality. Any additional features and functionalities will be left for future releases**

## Building, setup, and running
[!NOTE]
I've only ran/developed this on Windows (since it's my main dev platform, feel free to judge)

### Production

#### Building
Nothing special, clone the repo and hit it with `go build .`

[!NOTE]
You need to move this new found compiled build to another folder (unless you specified the location in the command), if not, the program will interfere with the included testing files

#### Setup

Before running the code, you must must create a specific file at the location where you are running the program: `./userFiles/links.json`.

`links.json` defines the folders to be linked, and their syncing parameters.

The general structure is:
```json

"links":[
    {
        "uuid": number,
        "syncDirection": "bi",
        "name": string,
        "folderA": string,
        "folderB": string,
        "folderRecursion": boolean,
        "syncOnStartup": true,
        "continuousSync": true,
        "syncType": "normal" 
        // Although there are other properties, as of writing, the following properties are the most generic and error-free
        // To view other properties, or other possible values for the listed properties, please view the `FolderLink` type in `folderlink.go`
    }
]

```

#### Running
Now that you have the the program built, and the `./userFiles/links.json` file setup, you can just run the program and enjoy.

### Development

The repo already includes a defined `./userFiles/links.json` file, along with some data files. So no need to worry about creating one and sourcing files.

#### Running

`go run .` and enjoy

## Program Structure

The code may seem convoluted *(mainly because it is, sorry for the spaghetti)*, but once you learn a few key terms and concepts, it makes more sense *(imo)*.

### Key Terms:

<ins>**FolderLink**</ins>: A `FolderLink` is simply an object that defines two folders that should be synced to each other. It also provides additional properties that both identify the `FolderLink` (eg. `UUID`) and also define how the folders should stay in sync (eg. `SyncType`). 

The name derives from the idea that you are 'linking' two folders together.

*A FolderLink is a direct representation of an object in `"links"` within the `./userFiles/links.json` file*

<ins>**WalkItem**</ins>: 

## FAQ

### What does it mean to be 'in-sync'?
Generally, 'in-sync' means that the targetted files between both folders are 1:1 with each other. 

However, this may not be what you desire all the time! Maybe you want the contents of one folder to be in a specific file type, while the other folder in another file type. While these files may house the same 'content', its actual representation is different.

*For example, one folder may house `.sav` files while the other houses `.srm`*

With this in mind, 'in-sync' now means that **the targetted files between both folders represent the same content.** (with the definition of the content being the users choice).

### What does FLS have to do with this?

***Folder-Link-Sync***  **acts as the manager for this process.**

The whole goal of **FLS** is automate this process of folder syncing. It allows you to define parameters of `What` files you want synced and `How` you'd like 'in-sync' to be defined.

### In other words, FLS aims to be:
- Lightweight
    - Can run quietly in the background without eating system services
- Customizable
    - Allows you to sync the way you'd like
- Zero maintenance
    - Once installed and set up, you can forget it

### Seems like a whole lot of talk, this codebase is disgusting and it's not that impressive

Ok yeah, fair, good point.

I originally made this as a quick and dirty solution of syncing the saves on my PSP with my computer/emulator (and vice-versa). 

When working on it for this need, I originally planned on making a program that syncs folders 1:1, as native PSP saves worked with the emulator I used *(PPSSPP)*. 
But, as soon as I worked on adding the functionality for other types of saves *(ie. saves from emulators for other consoles)*, I found issues with file formats and folder structures. 
For example, my PSP was using the `.sav` file extension for GBA games, while my PC used the `.srm` extension. Both of these formats handled the data the exact same; the extensions could be renamed *(without any sort of file conversion)* and work perfectly fine between devices. Since the data was the exact same, the file format became almost arbitrary.

When updating the logic *(and also getting lost in the sauce)* to fit my needs, I inadvertently planted the seeds for a syncing program that is flexible enough to fit any need.

Well, after I had it working for my use case, I abandoned the codebase and never let those seeds to grow. 

### Sounds like yap

shut up

