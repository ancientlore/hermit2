# ideas

* Have a browseable interface and implement browsers for file system, databases, processes, ...
* Cache browsed folders or things
* Embed https://github.com/containous/yaegi for extensions, even browseable things...

    type Browser interface {
        sort.Interface
        
        Enter(i int) (Browser, error)
        Reload()


    }