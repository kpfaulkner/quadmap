# quadmap
Quadtree (kind of) using quad indexes instead of true tree

Based off an idea from https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN

## Notes
Now includes the bulk of the non-quadkey data in sqlite database. This is to test
if the performance of shifting data to sqlite is bad enough to skip this experiment.


## PLAN

- Overall aim is to determine if AOI overlaps any of the tiles in the quadmap.
- Get cover quadkeys for AOI
- If we have a large quadkey ... then can we still check the quadmap for that quadkey even if the 
  real entry is for a smaller quadkey? (ie have zoom 18 but in the quadmap we have zoom 20). If we start
  search at larger quadkey 
- AOI is converted into cover quadkeys... then reduced to the smallest possible quadkey (22?)
- Search quadmap for the appropriate quadkeys... increasing quadkey depth until level 14?

## TODO

- Determine eviction/LRU policy. (evict tiles greater than a particular depth)
  - Will need to remove entries from Tile.groups
  - Reset watermark when removing particular entries
  

## MISC

- Get tile that covers AOI
- then get all surveys within that tile.... <- tricky bit.
- perform geom intersection