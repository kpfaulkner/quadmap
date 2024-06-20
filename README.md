# quadmap
Quadtree (kind of) using quad indexes instead of true tree

Based off an idea from https://learn.microsoft.com/en-us/bingmaps/articles/bing-maps-tile-system?redirectedfrom=MSDN


## TODO

- Determine eviction/LRU policy. (evict tiles greater than a particular depth)
  - Will need to remove entries from Tile.groups
  - Reset watermark when removing particular entries
  
