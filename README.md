# tcbscrape
Script to scrape tcbscans website. 

- 3200  Chapters
- 65737 Unique Images (Without modifications)

## Sizes (Estimate)
losless - 20M jxl, 35M png
lossy + resized - 5-10M jpg

64G png raws if each image takes 1Mb
32G losless jxl
20G losless jxl sans credits
10G low Quality

estimated 9GB of Credits in raws not really necessary
Could put credits into a database using smart ocr python stuff

## optimizer for

- output quality
- image format
    * oxipng (save 10%, quick)
    * oxipng-zopfil (most compatible, slowest)
    * lossless-jxl (40% efficient, quick)
    * lossless-webp (?)
    * lossless-avif (?)
    
    * lossy-avif (web safe)
    * lossy-webp (web safe)
    * lossy-jpeg (client safe)
    * lossy-jpeg-xl (least)

- resize options
    * 1080h
    * 720h

- storage format is jpeg xl

credit detection

      2 jpeg
  60502 jpg
   5233 png

