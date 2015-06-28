###

    OK, to explain this voodoo...
    Why didn't we stay with the requirejs optimizer? I'm glad you asked...
    * The requirejs optimizer was failing with text files, maps and other misc issues with libraries.
    * It also didn't give us fine-grained control over the bundles - it recursively included everything.
    * Finally, it took up to the 20 seconds to compile, every bloody time.

    In with the new...
    * When you run "gulp" now, we run through the list of bundles configured below and figure out a
        new name based on the timestamps of the modules. This also allows us to skip the packages that
        haven't changed.
    * We then inject the new requirejs config with bundles directly into the index.html (this may change).
    * We rebuild all 'dirty' bundles by concatenating all it's modules with the define syntax wrapped on it.

    Benefits:
        * We have full fine-grained control of how the bundling strategy will work
        * Any libraries not included here will be loaded one-by-one as normal so it's safe
        * It's blazing fast, even if all files are rebuilt.

    TODO:
        * Get the generated CSS to be a unique name based on timestamps and skip when already exists
        * Look at the app.conf - if testing "dist" folder mode, then reload automatically (Dev-Mode only)
        * Tidy up older files (something like delete files older than a week)

Requires:
    npm install --save-dev gulp-replace
    npm install --save-dev gulp-html-replace
    npm install js-string-escape
###

# Node modules
fs = require 'fs'
vm = require 'vm'
merge = require 'deeply'
chalk = require 'chalk'
es = require 'event-stream'
browserSync = require 'browser-sync'
getPort = require 'get-port'
changed = require 'gulp-changed'
jsStringEscape = require('js-string-escape')

# Gulp and plugins
gulp = require 'gulp'
$ = require('gulp-load-plugins')()

# Config
bases =
    app: './app/',
    dist: './dist/',
    src: './src/'
    docs: './docs/'
    public: './public/'

bases.bower = bases.src + 'libs/bower/'
bases.css = bases.src + 'css/'
bases.app = bases.src + 'app/'
paths =
    images: ['images/**/*.png']

requireJsRuntimeConfig = vm.runInNewContext fs.readFileSync('src/app/require.config.js') + ' require'

bundleInfo = null
requireLibsPathCache = {}
bundleDLMCache = {}# Bundles Date last modified cache
requireConfig = ''


###
This defines all bundles used by the app. These bundles load in one go are are therefore
much faster to load (browsers only load 8ish files at a time from a domain)
###
getBundlesInfoWithTimestamps = () ->
    ### These common libs problematic for some reason
        Probably because they are not AMD compatible modules
            'mousetrap'
            'mousetrap-global'
            'moment'
            'raven'
            'typeahead'
            'signals'
            'spin'
            'router'
            'text'
            'bindings-ladda'
            'bindings-misc'
    ###
    bundleConfig = {
        'redactor': [
            'redactor',
            'redactor-video',
            'redactor-table'
        ]
        'jquery': [
            'redactor',
            'redactor-video',
            'redactor-table'
        ]
        'libs': [
            'bindings-compareAt'
            'bootstrap'
            'bootstrap-select'
            'crossroads'
            'hasher'
            'jquery-mockjax'
            'jquery-mousewheel'
            'jquery-ui'
            'jquery.validate'
            'knockout'
            'knockout-mapping'
            'knockout-projections'
            'knockout-sortable'
            'fileupload'
            'clipboardPaste'
            'highcharts'
            'knockout-repeat'
            'knockout-switch-case'
            'punches'
            'ladda'
            'lodash'
            'lscache'
            'mCustomScrollbar'
            'modernizr'
            'calendar'
        ]
        'startup': [
            'lifevault'
            'config-components'
            'components': [
                'page-login'
                'page-reset-password'
                'page-forgot-password'
            ]
        ]
        'common': [
            'components': [
                'modal-imageupload'
                'modal-keyboard-shortcuts'
                'modal-prompt'
            ]
        ]
        'dashboard': [
            # Essential components
            'components': [
                'nav-main'
                'titlebar'
                'loadingIndicator'
                'widget-uiMessage'
                'page-loading'
                'page-dashboard'
                'section-blankslate'
                'modal-confirm'
            ]
        ]
        'journal': [
            'components': [
                'section-pagination'
                'page-journal'
            ]
        ]
        'myprofile': [
            'components': [
                'page-myprofile'
                'pageLHS-settings-nav'
                'section-user-profile'
            ]
        ]
        'search': [
            'components': [
                'page-search'
            ]
        ]
        'rare':[
            'components':[
                'page-errorLoadingMsg'
            ]
        ]


    }

    # Auto build any missing components into bundles
    # For every file in allComponents, check to see if it is already in a bundle
    # 1. Flatten components into an array
    componentsInBundles = []
    for bundleName, origBundleList of bundleConfig
        i = origBundleList.length-1
        while i > -1
            filePath = origBundleList[i]
            if typeof filePath is 'object' and typeof filePath.components isnt 'undefined'
                for component in filePath.components
                    if componentsInBundles.indexOf(component) isnt -1
                        console.error chalk.red("Error - component #{component} is in 2 bundles!")
                    else
                        componentsInBundles.push component
            i--
    extraBundleRequireList = []
    allComponents = vm.runInNewContext fs.readFileSync('src/app/config/components.js')
    for component in allComponents
        if componentsInBundles.indexOf(component) is -1
            #console.info "componentsInBundles.", component, componentsInBundles.indexOf(component)
            extraBundleRequireList.push component
    if extraBundleRequireList.length > 0
        $.util.beep()
        console.error chalk.bgRed(chalk.black "ERROR:") + chalk.red(" You have #{extraBundleRequireList.length} files not in bundles.")
        console.error chalk.bgBlue("'" + extraBundleRequireList.join("'\n'"))
        bundleConfig['ungrouped'] = [
            'components': extraBundleRequireList
        ]

    # HANDLE COMPONENT DEFINITIONS
    # Loop through each bundle, convert any components into def and html files
    for bundleName, origBundleList of bundleConfig
        i = origBundleList.length-1
        while i > -1
            filePath = origBundleList[i]
            if typeof filePath is 'object' and typeof filePath.components isnt 'undefined'
                components = filePath.components
                componentList = origBundleList.splice(i,1)
                for componentName in components
                    origBundleList.push('app/components/'+componentName+'/def')
                    try # The template.html file may not exists - not required for every component
                        stat = fs.statSync("#{bases.app}/components/#{componentName}/template.html")
                        origBundleList.push("text!app/components/#{componentName}/template.html")
                continue
            i--

    # Loop through each bundle, then all files, then find the latest timestamp
    # then modify the bundle name to reflect the latest timestamp
    startupFileAlreadyExisted = true
    bundleOrigNames = {}
    for bundleName, origBundleList of bundleConfig
        bundleList = origBundleList.slice(0)
        i = 0
        while i < bundleList.length
            filePath = bundleList[i]
            # Determine the correct file path
            finalFilePath = filePath
            if finalFilePath.indexOf('text!') == 0
                finalFilePath = finalFilePath.replace('text!','')
            else
                # Use path if defined in require.config.js
                lookupPath = requireJsRuntimeConfig.paths[filePath]
                if lookupPath?
                    finalFilePath = lookupPath
                finalFilePath = finalFilePath + ".js"
            finalFilePath = bases.src + finalFilePath
            # Cache the file path
            requireLibsPathCache[filePath] = finalFilePath
            # Get the date last modified of the file
            stat = fs.statSync( finalFilePath )
            if bundleDLMCache[bundleName] is undefined or stat.mtime >= bundleDLMCache[bundleName]
                bundleDLMCache[bundleName] = stat.mtime
            i++

        bundleFinalName = bundleName + "-" + bundleDLMCache[bundleName].getTime()
        bundleConfig[bundleFinalName] = bundleConfig[bundleName]
        delete bundleConfig[bundleName]
        bundleOrigNames[bundleFinalName]= bundleName



    return {
        bundles: bundleConfig
        bundleOrigNames: bundleOrigNames
    }






gulp.task 'html', ['css','requirejsBundling'], ->

    # Check if config file exists, or assume it's production
    try
        stat = fs.statSync( bases.public + 'config.js' )
        requireConfig = '<script src="public/config.js"></script>\n' + requireConfig


    # Next we want to rebuild the index.html to include the right js file
    return gulp.src bases.src + 'index.html'
            .pipe $.htmlReplace(
                'css': ['dist/' + finalCSSFile]
                'debug': []
                'note': ''
                'requireConfig': requireConfig
            )
            .pipe gulp.dest bases.dist



# Discovers all AMD dependencies, concatenates together all library .js files
gulp.task 'requirejsBundling', ['coffee'], ->
    bundleInfo = getBundlesInfoWithTimestamps()

    # For each bundle in bundleInfo
    # Merge them together can create the output
    # The get the requirejs file and append in the new bundle config

    numBundles = 0
    requireConfig = fs.readFileSync bases.app + 'require.config.js'
    requireConfig = '<script>\n' + requireConfig + '</script>'

    bundlesConfig = 'require.config({\n"bundles":{'
    for bundleName, fileList of bundleInfo.bundles
        numBundles++
        bundlesConfig += '\t'
        if numBundles > 1 then bundlesConfig += ','
        bundlesConfig += '"dist/' + bundleName + '":[\n'
        for i, filePath of fileList
            if i > 0 then bundlesConfig += ',\n'
            bundlesConfig += '\t\t"' + filePath + '"'
        bundlesConfig += '\n\t]\n'
    bundlesConfig += '\n}});'

    requireConfig += '\n<script src="libs/bower/requirejs/require.js"></script>'
    requireConfig += '\n<script src="libs/bower/jquery/dist/jquery.js"></script>'
    requireConfig += '\n<script>\n' + bundlesConfig + '\nrequire(["app/init"]);\n</script>'


    # Concat files in each bundle together
    for finalBundleName, requirePaths of bundleInfo.bundles
        bundleName = bundleInfo.bundleOrigNames[finalBundleName]

        # If the final file already exists, then we skip it, we have it compiled already
        finalBundleFullPath = bases.dist + bundleName + "-" + bundleDLMCache[bundleName].getTime() + ".js"
        finalBundleFileAlreadyExists = false
        try
            stat = fs.statSync( finalBundleFullPath )
            finalBundleFileAlreadyExists = true
            console.info "SKIPPING EXISTING BUNDLE", finalBundleFullPath
        catch
            console.info "*** BUILDING NEW BUNDLE ***", finalBundleFullPath

        # Skip rebuilding the file if we already have it
        if finalBundleFileAlreadyExists then continue

        # Build the file
        buf = ''
        for i, requirePath of requirePaths
            if i > 0 then buf += '\n\n'
            fileData = fs.readFileSync requireLibsPathCache[requirePath]
            fileData = fileData.toString()
            if requirePath.indexOf(".html") == -1
                if fileData.indexOf('define(') > -1
                    buf += fileData.replace('define(','define("'+requirePath+'",')
                else
                # There is no existing define so just include the script
                # and do the empty define manually
                    buf += fileData
                    buf += '\ndefine("'+requirePath+'", function(){});'
            else
                buf += "define('"+requirePath+"',[],function(){return '"
                buf += jsStringEscape(fileData)
                buf += "'});"

        # Save the file
        bundleFileName = bases.dist + finalBundleName + '.js'

        ugly = (bundleFileName) -> # Done in fn becaise bundleFileName gets overwritten in loop
            fs.writeFile bundleFileName, buf, (err) ->
                if err then return console.log(err)
                # Minify the new bundle file
                # Done on a seperate thread so we're not waiting for it
                console.info "UGLY ", bundleFileName
                gulp.src bundleFileName
                    .pipe $.uglify
                        preserveComments: 'none'
                    .pipe gulp.dest bases.dist
                return
        ugly(bundleFileName)


    return




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
###
        .pipe $.uglify
            preserveComments: 'none'
###

# Concatenates CSS files, rewrites relative paths to Bootstrap fonts, copies Bootstrap fonts
finalCSSFile = ''
gulp.task 'css', ->

    fileList = fs.readdirSync(bases.css)
    for file in fileList
        stat = fs.statSync "#{bases.css}#{file}"
        if @timestamp is undefined or stat.mtime >= @timestamp
            @timestamp = stat.mtime

    finalCSSFile = 'css-' + @timestamp.getTime() + ".css"
    finalCSSFullPath = bases.dist + finalCSSFile
    finalCSSFileAlreadyExists = false
    try
        stat = fs.statSync( finalCSSFullPath )
        finalCSSFileAlreadyExists = true
        console.info "SKIPPING EXISTING CSS FILE", finalCSSFullPath
    catch
        console.info "*** BUILDING NEW CSS ***", finalCSSFullPath

    if finalCSSFileAlreadyExists then return

    appCss = gulp.src [bases.css+'*.css', bases.public+'upload/css/jquery.fileupload.css', bases.public+'public/css/card.css']
    combinedCss = es.concat appCss
        .pipe $.concat finalCSSFile

    return es.concat combinedCss
        .pipe gulp.dest bases.dist



# Removes all files older than 1 hour ago from ./dist/
gulp.task 'clean', ->
    cutoff = new Date() - (1000 * 60 * 60) * 1
    fileList = fs.readdirSync(bases.dist)
    for file in fileList
        stat = fs.statSync "#{bases.dist}#{file}"
        if stat.atime < cutoff # Note the use of atime (Access Time) instead of mtime (Modified time)
            fs.unlink "#{bases.dist}#{file}", (err) ->
                if err then console.error('Error deleting', file)
                else console.log('successfully deleted', file)
    return

# The default task is compile which generates the merged css,
# and also generates an updated index.html via
# html -> js -> requirejsBundling -> coffee
gulp.task 'compile', ['html'], (callback) ->
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


gulp.task 'watch', ['html'], ->
    browserSync
            port: 50243
            ghostMode: false

    gulp.watch([bases.app + '**/*.coffee'], ['coffee']).on('change', browserSync.reload)
    gulp.watch([bases.css + '**/*.css']).on('change', browserSync.reload)
    gulp.watch([bases.app + '**/*.html']).on('change', browserSync.reload)

gulp.task 'default', ['compile']