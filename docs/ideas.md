# ideas

* Have a browseable interface and implement browsers for file system, databases, processes, ...
* Cache browsed folders or things
* Embed https://github.com/containous/yaegi for extensions, even browseable things...

    type Browser interface {
        sort.Interface
        
        Enter(i int) (Browser, error)
        Reload()


    }

TODO:

- Configure shell that launches with $
- Add themes and config for them
    - Greyscale
    - Amber monitor
    - Green monitor
    - Other common styles (light, dark, high contrast)
- Add configurable commands (gscript)
    - What do the variables look like?
- Add more browsable type (s3, etc.)
- Add "workspaces" which used to be bookmarks
- Search feature in the text viewer / n p
