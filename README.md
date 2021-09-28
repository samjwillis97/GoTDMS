# GoTDMS

GoTDMS is a CLI tool written in Go to Read TDMS Files produced by LabVIEW. A lot of this project is based off of the Python project with a similar goal npTDMS (https://github.com/adamreeve/npTDMS)

## Primary Functionality

- List File as Tree
  - tdms list [file]
- List Groups
  - tdms list groups [file]
- List Channels
  - tdms list channels [group] [file]
- List Properties of Group/Channel as Strings
  - tdms list properties [group] [channel] [file]
- Get Property of Group/Channel with a cast
  - tdms read property [group] [channel] [name] [type]
- Output Group/Channel/s with Offset + Length
  - tdms read data

## TODO

- Split Project up into smaller files, more manageable
- Split commands up into different files/commands
  - tdms list groups -> tdms-list groups
  - tdms list channels -> tdms-list channels
  - tdms list properties -> tdms-list properties
  - tdms read -> tdms-read
  - Add a tdms-plot for fun
