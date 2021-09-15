# GoTDMS

GoTDMS is a CLI tool written in Go to Read TDMS Files produced by LabVIEW. A lot of this project is based off of the Python project with a similar goal npTDMS (https://github.com/adamreeve/npTDMS)

## Primary Functionality

- List Groups
  - tdms list groups <file>
- List Channels
  - tdms list channels <group> <file>
- List Properties of Group/Channel as Strings
  - tdms list properties <group> <channel> <file>
- Get Property of Group/Channel with a cast
  - tdms read property <group> <channel> <name> <type>
- Output Group/Channel/s with Offset + Length
  - tdms read data 

## Data Type Definitions from npTDMS

### TDMS Segment

Contains:

- Position
    - Starting Position, on the first segment, = 0
- Number of Chunks
    - from Calculate Chunks
- Ordered Objects []:
    - Path
    - Number Values
    - Data Size
    - Has Data?
    - Data Type
- ToC Mask
- Next Segment Position
  - Determined from Lead In
- Data Position
    - Determined from Lead In
    - Segment Position + Lead Size (28) + Raw Data Offset
- Final Cunk Lengths Override ??
    - from Calculate Chunks
- Object Index ??

