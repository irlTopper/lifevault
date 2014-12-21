# Node modules
fs = require 'fs'
vm = require 'vm'
merge = require 'deeply'
chalk = require 'chalk'
es = require 'event-stream'
browserSync = require 'browser-sync'
getPort = require 'get-port'
changed = require 'gulp-changed'

# Gulp and plugins
gulp = require 'gulp'
$ = require('gulp-load-plugins')()

# Config
bases =
    app: './app/',
    dist: './dist/',
    src: './src/'
    docs: './docs/'

bases.bower = bases.src + 'libs/bower/'
bases.css = bases.src + 'css/'
bases.app = bases.src + 'app/'

paths =
    images: ['images/**/*.png']

requireJsRuntimeConfig = vm.runInNewContext fs.readFileSync('src/app/require.config.js') + ' require'
requireJsOptimizerConfig = merge requireJsRuntimeConfig,
    out: 'scripts.js'
    baseUrl: bases.src
    name: 'app/startup'
    paths:
        requireLib: 'libs/bower/requirejs/require'
    include: [
        'requireLib'
        'app/components/nav-main/def'
        'app/components/page-dashboard/def'
        'app/components/titlebar/def'
    ]
    insertRequire: ['app/startup']
#    bundles:
    # If you want parts of the site to load on demand, remove them from the 'include' list
    # above, and group them into bundles here.
    # 'bundle-name': [ 'some/module', 'another/module' ],
    # 'another-bundle-name': [ 'yet-another-module' ]

gulp.task 'coffee', ->
    return gulp.src(bases.app + '**/*.coffee',
        base: bases.app
    )
        .pipe changed bases.app, { extension: '.js' }
        .pipe $.plumber
            errorHandler: console.log
        .pipe $.coffeelint './node_modules/teamwork-coffeelint-rules/coffeelint.json'
        .pipe $.coffeelint.reporter()
        .pipe $.coffeelintThreshold 10, 0, (numberOfWarnings, numberOfErrors) ->
            $.util.beep()
            console.error chalk.bgRed(chalk.black "ERROR:") + chalk.red(" CoffeeScript compilation failure") + " due to $.coffeeLint violations; see above. Warning count: #{chalk.blue numberOfWarnings}. Error count: #{chalk.red numberOfErrors}."
        .pipe $.coffee
            bare: true
        .pipe gulp.dest bases.app


# Discovers all AMD dependencies, concatenates together all required .js files, minifies them
gulp.task 'js', ['coffee'], ->
    return $.requirejsBundler requireJsOptimizerConfig
        .pipe $.uglify
            preserveComments: 'some'
        .pipe gulp.dest bases.dist

# Concatenates CSS files, rewrites relative paths to Bootstrap fonts, copies Bootstrap fonts
gulp.task 'css', ->
    bowerCss = gulp.src bases.bower + 'components-bootstrap/css/bootstrap.min.css'
        .pipe $.replace /url\((')?\.\.\/fonts\//g, 'url($1fonts/'

    appCss = gulp.src bases.css + '*.css'

    combinedCss = es.concat bowerCss, appCss
        .pipe $.concat 'css.css'

    fontFiles = gulp.src bases.bower + 'components-bootstrap/fonts/*',
        base: bases.bower + 'components-bootstrap/'

    return es.concat combinedCss, fontFiles
        .pipe gulp.dest bases.dist

# Copies index.html, replacing <script> and <link> tags to reference production URLs
gulp.task 'html', ->
    return gulp.src bases.src + 'index.html'
        .pipe $.htmlReplace(
            'css': 'css.css'
            'js': ['scripts.js', 'config.js']
        )
        .pipe gulp.dest bases.dist


# Imagemin images and ouput them in dist
gulp.task 'imagemin', ->
    return gulp.src paths.images
        #.pipe $.imagemin()
        .pipe gulp.dest bases.dist

# Removes all files from ./dist/
gulp.task 'clean', ->
    return gulp.src bases.dist + '*'
        .pipe $.clean()

gulp.task 'clean-docs', ->
    return gulp.src bases.docs + '*'
        .pipe $.clean()

gulp.task 'generate-docs', ['clean-docs'], ->
    return $.biscotto()
        .pipe gulp.dest bases.docs

gulp.task 'docs', ->
    return gulp.src bases.docs + 'index.html'
        .pipe $.open()

gulp.task 'compile', ['html', 'js', 'css', 'imagemin'], (callback) ->
    callback()
    console.log '\nPlaced optimized files in ' + chalk.magenta 'dist/\n'


gulp.task 'open', ['compile'], ->
    getPort (err, port) ->
        browserSync
            server:
                baseDir: [
                    bases.dist
                    bases.app
                    bases.src
                ],
            port: port

gulp.task 'watch', ['compile'], ->
    gulp.watch bases.src + 'index.html', ['html', browserSync.reload]
    gulp.watch bases.src + 'app/*.js', ['js', browserSync.reload]
    gulp.watch bases.app + '**/*.coffee', ['js', browserSync.reload]
    gulp.watch bases.css + '*.css', ['css', browserSync.reload]
    gulp.watch bases.bower + '**', ['js', 'css', browserSync.reload]
    gulp.watch paths.images, ['imagemin', browserSync.reload]

gulp.task 'coffeewatch', ->
    browserSync
            port: 50243
            ghostMode: false

    gulp.watch bases.app + '**/*.coffee', ['coffee', browserSync.reload]
    gulp.watch bases.app + '**/*.html', browserSync.reload


gulp.task 'default', ['compile']