define ['moment', 'knockout'], (moment, ko) ->

    Utility = ->

    Utility::IsEmail = (email) ->
        regex = /^([a-zA-Z0-9_.+-])+\@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$/
        return regex.test(email)

    Utility::IsValidDomain = (domain) ->
        re = new RegExp(/^((?:(?:(?:\w[\.\-\+]?)*)\w)+)((?:(?:(?:\w[\.\-\+]?){0,62})\w)+)\.(\w{2,6})$/)
        domain.match re

    Utility::CalculatePercentageChange = (current, previous) ->
        return 100 if current is 0
        current = ko.utils.unwrapObservable(current)
        previous = ko.utils.unwrapObservable(previous)
        Math.floor(((current - previous) / current) * 100)

    Utility::MinutesIntToString = (minutes, abbreviate = false) ->
        minutes = ko.utils.unwrapObservable(minutes)
        return unless minutes?

        minutesStr = if abbreviate then "m" else " minutes"
        hoursStr = if abbreviate then "h" else " hours"
        daysStr = if abbreviate then "d" else " days"

        if minutes < 60
            return "#{Math.ceil(minutes)}#{minutesStr}"

        hours = Math.floor(minutes / 60)
        minutes = Math.round(minutes % 60)
        if hours > 24
            days = Math.floor(hours / 24)
            hours = hours % 24

            return "#{days}#{daysStr} #{hours}#{hoursStr} #{minutes}#{minutesStr}"

        return "#{hours}#{hoursStr} #{minutes}#{minutesStr}"

    Utility::convertZuluStringToMoment = (dateStr) ->
        return moment.min( moment( dateStr, 'YYYY-MM-DDTHH:mm:ssZ'), moment())

    # There is PROBABLY a better way to do this.  If you know of it, go for it :)
    Utility::PrefixKeys = (prefix, object) ->
        keys = {}
        for key in Object.keys(object)
            keys["filters.#{key}"] = object[key]
        return keys

    Utility::SafeHTML = (val) ->
        return String(val).replace(/&/g,"&amp;").replace(/\"/g,"&quot;").replace(/\'/g,"&#39;").replace(/</g,"&lt;").replace(/>/g,"&gt;")

    Utility::URLSafeString = (val) ->
        return String(val).replace(/\s/g, '-').replace(/[^a-zA-Z0-9-_]/g, '').toLowerCase()


    ###
    Given a file size, returns a human friendly text string such as "27KB" for "3.2 Mb (medium)"
    use like tw.getFriendlyBytes( 888, {getHint:true} );
    @method getFriendlyBytes
    @param {Number} The size of a file.
    ###
    Utility::GetFriendlyBytes = (size, opts) ->
        i = 0
        local_units = [
            "B"
            "KB"
            "MB"
            "GB"
            "TB"
        ]
        size = parseInt(size, 10)
        while size >= 1024
            size /= 1024
            ++i

        #Use the Intl.NumberFormat to format where it's supported
        if window.Intl and window.Intl.NumberFormat
            unless @IntlNumberFormatter?
                @IntlNumberFormatter = new window.Intl.NumberFormat("en-US", maximumFractionDigits: 2)
            local_result = @IntlNumberFormatter.format(size)
        else
            local_result = (if i is 0 then size else size.toFixed(1))
        local_result += local_units[i]
        return local_result


    ###
    Returns the number with commas eg "5673" -> "5,673"
    @method NumberFormat
    @param {Number} The size of a file.
    ###
    Utility::NumberFormat = (num) ->
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",")

    # Using a single point of access for scheme will allow us to make
    # future changes to the requirements for loading over https (custom domains, etc)
    # without much hassle.
    Utility::Scheme = ->
        if twDeskConfig.isDeveloperMode then return "http://"
        if document.URL.indexOf("teamwork.com") > -1 or document.URL.indexOf("teamworkpm.net") > -1
            return "https://"
        return "http://"

    ###
    Takes care of the need for either a ? or & when extending a url
    use like url = extendURL(url,'key=val');
    @method ExtendURL
    @param {String} The url to extend.
    @param {Object} Param pair. The bit to append to the url. should look like 'x=y'}
    ###
    Utility::ExtendURL = (url,str) ->
        url += if url.indexOf("?")==-1 then '?' else '&'
        if typeof str is "string"
            url += str
        else if typeof str is "object"
            params = $.param( str )
            url += params
        return url

    ###
    Fins a parameter on the query string and returns
    @method GetParamByName
    @param {String} The query param name to get
    ###
    Utility::GetQueryStringParamByName = (name) ->
        name = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]')
        regex = new RegExp('[\\?&]' + name + '=([^&#]*)')
        results = regex.exec(location.search)
        if results == null then null else decodeURIComponent(results[1].replace(/\+/g, ' '))


    Utility::GetAttachmentIcon = (filename, mimetype) ->
        extension = undefined
        icons =
            outlook: ["msg"]
            doc: [
                "odt"
                "rtf"
                "txt"
                "wpd"
                "wps"
                "doc"
                "docx"
            ]
            text: ["txt","text"]
            pdf: ["pdf"]
            latex: ["tex"]
            csv: ["csv"]
            data: [
                "dat"
                "efx"
                "epub"
                "ged"
                "ibooks"
                "sdf"
                "tax2010"
                "xml"
                "log"
            ]
            presentation: ["key"]
            ppt: [
                "pps"
                "ppt"
                "pptx"
            ]
            vcard: [
                "vcf"
                "vcard"
            ]
            script: ["json","js"]
            audio: [
                "aif"
                "aiff"
                "iff"
                "m3u"
                "m4a"
                "mid"
                "mp3"
                "mpa"
                "ra"
                "wav"
                "wma"
                "ogg"
            ]
            video: [
                "3g2"
                "3gp"
                "asf"
                "asx"
                "avi"
                "mov"
                "mp4"
                "mpg"
                "rm"
                "srt"
                "vob"
                "wmv"
                "mkv"
            ]
            flash: [
                "fl"
                "fla"
                "swf"
                "flv"
            ]
            threed: [
                "3dm"
                "max"
                "obj"
            ]
            raster: [
                "bmp"
                "dds"
                "gif"
                "jpg"
                "jpeg"
                "png"
                "pspimage"
                "tga"
                "thm"
                "tif"
                "tiff"
                "yuv"
            ]
            ps: ["psd"]
            ai: ["ai"]
            fw: ["fw"]
            ae: ["ae"]
            keynote: ["keynote"]
            numbers: ["numbers"]
            pages: ["pages"]
            pr: ["pr"]
            vector: [
                "eps"
                "svg"
                "cgm"
            ]
            xls: [
                "xlr"
                "xls"
                "xlsx"
            ]
            sql: ["sql"]
            access: [
                "accdb"
                "mdb"
            ]
            database: [
                "db"
                "dbf"
                "pdb"
                "sqlite"
            ]
            macos: [
                "app"
                "dmg"
            ]
            windows: [
                "exe"
                "com"
                "gadget"
                "bat"
                "wsf"
            ]
            linux: [
                "sh"
                "rpm"
            ]
            scriptcode: [
                "asp"
                "aspx"
                "cfm"
                "js"
                "jsp"
                "cfc"
                "pl"
                "py"
            ]
            documentcode: [
                "css"
                "c"
                "class"
                "cpp"
                "cs"
                "dtd"
                "java"
                "m"
            ]
            feed: ["rss"]
            php: ["php"]
            webdoc: [
                "html"
                "htm"
                "xhtml"
            ]
            archive: [
                "zip"
                "7z"
                "deb"
                "gz"
                "pkg"
                "rar"
                "sit"
                "sitx"
                "tar.gz"
                "zipx"
                "bz2"
                "jar"
            ]
            disc: [
                "bin"
                "cue"
                "iso"
                "toast"
                "vcd"
            ]

        extension = filename.toLowerCase().split(".")
        extension = extension[extension.length - 1]
        for iconName of icons
            return (iconName + ".png")  if icons[iconName].indexOf(extension) isnt -1
        "default.png"
    ###
    Utility::ScrollIntoViewIfNeeded = (elm) ->
        jSelf = $(elm)
        if jSelf.length is 0 then return
        scrollTop = 0
        if document.documentElement and document.documentElement.scrollTop
            scrollTop = document.documentElement.scrollTop
        else
            scrollTop = document.body.scrollTop
        if document.viewport.getDimensions().height + scrollTop < jSelf.offset().top + jSelf.height()
            elm.scrollIntoView(false)
        return
    ###
    return Utility