HipChat Output for Mozilla Heka

An Atlassian HipChat Output for Mozilla's Heka.

# Installation

Create or add to the file {heka_root}/cmake/plugin_loader.cmake
```
add_external_plugin(git https://github.com/christianvozar/heka-hipchat-output master)
```

Then build Heka per normal instructions for your platform.

