# Lemur

A templating helper library

## TODO:

- [ ] test for, and handle non-existant tmplDir
- [ ] Add test where there is no _defaults/_index.html but there is another _defaults
      file



## Theme directory structure

Directories starting with an '_' underscore are known, expected, files.

theme_name/
├── layouts/
│   ├── _defaults/
│   │   ├── _index.html
│   │   ├── author-by-line.html
│   │   ├── date.html
│   │   └── figure.html
│   ├── template_name/
│   │   └── _main.html
│   ├── second_template_name/
│   │   ├── _index.html
│   │   └── my-rando-file.html
│   └── _public/
└── Readme.md

## Minimal theme

The most minimal theme is one that is just the defaults

theme_name/
└── layouts/
    └── _defaults/
        └── _index.html.tmpl
