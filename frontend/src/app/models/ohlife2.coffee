define [
    'lscache',
    'router',
    'knockout',
    'knockout-mapping',
    'userModel',
    'ladda',
    'moment',
    'knockout-switch-case',
    'knockout-repeat',
    'bootstrap-select'
],
(lscache, router, ko, mapping, userModel, Ladda, moment) ->

    ohlife2 = ->
        # ensure HTTPS when in production mode
        if not ohlife2Config.isDeveloperMode and window.location.protocol != "https:"
            window.location.href = "https:" + window.location.href.substring(window.location.protocol.length)
            return
        @ko = ko
        @mapping = mapping
        @lscache = lscache
        @essentialDataIsLoaded = ko.observable(false)
        @mainHolderClass = ko.observable('')
        @isLoading = ko.observable(false)
        @pageLoadingErrorXHR = ko.observable(null)
        @pageComponentToShow = ko.observable('')
        @router = router
        @hasher = router.hasher
        @currentRoute = router.currentRoute
        @currentRoute.subscribe (newRoute) => @onUpdateRoute(newRoute)
        @uiMessage = ko.observable null #used for displaying a message to the user
        @modals = ko.observableArray() #used or tracking loaded modals components

        # We use the lscache to remember if we were logged in before this reload
        # This is secure because the server side is really in charge and the first
        # ajax load will fail if the user isn't logged in.
        # The purpose of this is basically to save us having to do GET /login.json
        # whenever the user does a full reload. ps. If this fails, it's null = not logged in.
        savedLoggedInUser = lscache.get("loggedInUser")
        if savedLoggedInUser isnt null then @loggedInUser = mapping.fromJS(savedLoggedInUser)
        else @loggedInUser = null



        # Pretend the navigation just changed to kick things off
        @onUpdateRoute(@currentRoute()) #initial call needed because subscription not used yet


        $.ajaxPrefilter ( options ) ->
            if not options.beforeSend?
                options.beforeSend = (xhr) ->
                    xhr.setRequestHeader('twDeskVer', '0.2');



        # Some messages to save us having them everywhere in code
        @Msgs = {
            SessionExpired: "Sorry it looks like your session has expired.",
            AccessDenied: "The server has denied you access. Your permissions may have been changed. If this continues try logging out and back in again.",
            NetworkTimeout: "<strong>Network timeout</strong> &mdash; Please check your internet connection and try again. Email support@teamworkdesk.com if you still have problems.",
            ServerRejected: "<strong>Server rejected</strong> &mdash; Sorry the server rejected this request.",
            ServerRejectedWithMessage: "The server rejected this request with the message:",
            ServerError: "<strong>Server error</strong> &mdash; Sorry the server ran into a problem processing this request.",
        }

        ###
        CUSTOM KNOCKOUT BINDING HANDLERS HERE
        ###

        # See ladda.js - animations for submit buttons
        # This sets it up to be very eay to use with knockout
        ko.bindingHandlers.ladda = {
            init: (element, valueAccessor) ->
                l = Ladda.create(element)
                ko.computed({
                    read: () ->
                        state = ko.unwrap(valueAccessor())
                        if state
                            l.start()
                        else
                            l.stop()
                    ,
                    disposeWhenNodeIsRemoved: element
                })
        }


        # Redactor plugin drop downs - save us setting it up all the time
        @setupRedactorPlugins()

        # Browser Sync
        if ohlife2Config.isDeveloperMode then @InitBrowserSync()

        # Load the extra CSS and javascript
        @InitExtraCSSAndJavascript()

        return @


    # Using a single point of access for scheme will allow us to make
    # future changes to the requirements for loading over https (custom domains, etc)
    # without much hassle.
    ohlife2::Scheme = ->
        if ohlife2Config.isDeveloperMode then "http://" else "https://"

    ohlife2::Https = ->
        @Scheme().indexOf("https") isnt -1
    ohlife2::InitBrowserSync = () ->
        scriptPath = "//" + location.hostname + ":50243/browser-sync-client.1.3.7.js"
        jQuery.getScript(scriptPath)

    # Here we load CSS and scripts that aren't essential for the initial load but needed
    ohlife2::InitExtraCSSAndJavascript = () ->
        h = $('head')
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/bower/Ladda/dist/ladda-themeless.min.css') )
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/bower/jquery-minicolors/jquery.minicolors.css') )
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/redactor/redactor.css') )



    ohlife2::loadEssentialData = ( callback ) ->
        @InitSocket()

        ## This awkward looking syntax loads the json in parallel and only when it's finished
        ## toggles essentialDataIsLoaded to true.
        ## Todo: Handle network errors elegantly, maybe switch to the network error page
        @essentialDataIsLoaded true
        # Perform the callback - this will show the requested logged in page
        callback()

        return

    ohlife2::redirectToLogin = ( callback ) ->
         # Redirect to login but remember this page
        if app.currentRoute().page isnt "login"
            app.requestedRouteBeforeLoginRedirect = app.currentRoute()
            app.GoTo "login"
        return false

    ohlife2::init = (opts) ->
        ## Start the application by initializing the router and root binding
        @router.init()
        ko.applyBindings {}

    # For navigating around the app
    # Pass the silent flag to change it without the app updating
    # Handy for when something is added.
    ohlife2::GoTo = (hash,silent=false) ->
        console.assert( hash.indexOf("http") is -1, "Don't use GoTo for full URLs" )

        hash = hash.toLowerCase()

        if not silent
            @hasher.setHash(hash)
        else
            @hasher.changed.active = false
            @hasher.setHash(hash)
            @hasher.changed.active = true
        return

    ohlife2::onUpdateRoute = (newRoute) ->
        # Load the essential data asap - logged in pages won't show until this is done
        if newRoute.isLoggedInPage and not @essentialDataIsLoaded()
            @pageComponentToShow( 'page-loading' )
            @loadEssentialData () =>
                @pageComponentToShow( 'page-' + newRoute.page )
        else
            @pageComponentToShow( 'page-' + newRoute.page )

        # We store a userPageStatus for each route, always starts with "viewing"
        @currentRoute().userPageStatus = "viewing"

        if newRoute.isLoggedInPage
            if @socket then @socket.NotifyNewRoute()

        return


    ohlife2::FindUserById = (userId) ->
        match = ko.utils.arrayFirst app.users(), (user) =>
            return user.id() is userId
        return match

    ohlife2::FindInboxById = (inboxId) ->
        if inboxId is 0 then return @unifiedInbox
        match = ko.utils.arrayFirst @inboxes(), (inbox) =>
            return inbox.id() is inboxId
        return match

    ohlife2::GetTicketStatusNameByCode = (code) ->
        match = ko.utils.arrayFirst @ticketstatuses, (status) =>
            return status.code is code
        return if match? then match.name else "Unknown"

    # Will display a message to the user - a type can be passed in options
    # valid "type"'s' are 'error', 'warning', 'success' and 'info'
    ohlife2::FlashMessage = (msg, inOpts) ->
        opts = {
            text: msg,
            type: "info",
            timer: 3000
        }
        if inOpts?
            $.extend(opts, inOpts)

        app.uiMessage(opts)
        return

    # Cnvenience wrapper for FlashMessage with type set to error
    ohlife2::FlashSuccess = (msg, inOpts) ->
        opts = {
            type: "success"
            timer: 800
        }
        if inOpts?
            $.extend(opts, inOpts)
        @FlashMessage(msg, opts)
        return

    # Cnvenience wrapper for FlashMessage with type set to error
    ohlife2::FlashError = (msg, inOpts) ->
        opts = {
            type: "error"
            timer: 4000
        }
        if inOpts?
            $.extend(opts, inOpts)
        @FlashMessage(msg, opts)
        return


    ###
    MissingEssentialData() is used throughout the app when we try to access some
    essential data (see loadEssentialData) and it is missing.
    It could happen if try to access a newly added object or the user is messing
    about with the URL.
    We handle the missing data by first reloading everything... this should cause
    the essential data to reload - crude but effective.
    However if the data still doesn't load, we know the user is probably trying to
    access something invalid or that he doesn't has permissions for so we redirect.
    ###
    ohlife2::MissingEssentialData = (params) ->
        # If we've tried the reload already and still no luck, then go to the dash
        if lscache.get("MissingEssentialData-tryingReload") isnt null
            lscache.set "MissingEssentialData-tryingReload", null
            # app.GoTo "dashboard"
            app.FlashError("We couldn't load the page you requested &mdash; you might be denied access.")
            return

        lscache.set "MissingEssentialData-tryingReload", true, 1000
        location.reload()
        return



    # Use this is you were trying to load essential data for a page and there was a problem,
    # We will show an generic error message page and provide a try again button.
    ohlife2::showErrorLoadingMsg = (xhr) ->
        @pageLoadingErrorXHR(xhr)
        return

    # Use this if there was an error performing an action
    ohlife2::HandleAjaxError = (xhr) ->
        explOpts = ["Whoops!","Yikes!","Blimey!","Oh no!","Oh Dear!"]
        expl = explOpts[Math.floor(explOpts.length*Math.random())]
        if xhr.status == 401 # Expired session
            app.FlashError( app.Msgs.SessionExpired + " &ndash; <a href=\"javascript:document.location.reload()\">Login again to continue</a>" )
        else if xhr.status == 403 # Access Denied
            app.FlashError( app.Msgs.AccessDenied )
        else if xhr.status == 400 # Server rejected
            if xhr.responseJSON? and xhr.responseJSON.message
                @FlashError( "<strong>"+expl+"</strong> " + xhr.responseJSON.message )
            else if xhr.responseJSON? and xhr.responseJSON.errors and xhr.responseJSON.errors[0]
                @FlashError( "<strong>"+expl+"</strong> " + xhr.responseJSON.errors[0] )
            else
                @FlashError( app.Msgs.ServerRejected )
        else if xhr.status == 0 # Network timeout
            @FlashError( app.Msgs.NetworkTimeout )
        else
            if xhr.responseJSON? and xhr.responseJSON.message
                @FlashError( "<strong>"+expl+"</strong> &mdash; " + xhr.responseJSON.message )
            else if xhr.responseJSON? and xhr.responseJSON.errors
                @FlashError( "<strong>"+expl+"</strong> &mdash; " + xhr.responseJSON.errors[0] )
            else
                @FlashError( app.Msgs.ServerError )

        return


    # Wrapper function for showing a modal
    ohlife2::Confirm = (title, question, callback) ->
        app.ShowModal("confirm",{
            title: title
            question: question
            callback: callback
        },this)
        return

    # Loads a modal by adding it to the app.modals observable array
    # Also ensures a modal is not loaded twice
    ohlife2::ShowModal = (name,params={},holder) ->

        # Create something to hold this reference
        ModalRef = (name,params) ->
            @name = name
            @params = params
            @isLoaded = false
            @instance = params
            return @

        modalRef = new ModalRef(name,params)

        # Setup the loaded callback - modals are expected to call this
        $.extend(params, {
            modalId: app.modals().length+1
            modalRef: modalRef
        })

        app.modals.push modalRef
        return


    ohlife2::ModalInit = (modalId,modalVM,params={}) ->
        modalVM.templateLoaded = app.ko.observable(false)
        modalVM.modalDivId = modalId + params.modalId
        modalVM.modalRef = params.modalRef
        modalVM.allowFade = ko.observable(true)#used to turn off fade-in/out effect class


        ShowModal = (modalVM) =>
            modalVM.modalDiv.modal("show")
            app.currentModalVM = modalVM

            # Setup the after-closing-animation function:
            AttachOnhide(modalVM)

        AttachOnhide = (modalVM) =>
            modalVM.modalDiv.on('hidden.bs.modal', (e) ->

                # Clear the binding - not having this causes madness - Topper
                $( modalVM.modalDiv ).unbind( 'hidden.bs.modal' )

                # If not stacked, just close
                if not modalVM.stackedOnModalVM?
                    app.currentModalVM = undefined
                else# ..Otherwise restore the modal this was stacked on
                    ShowModal(modalVM.stackedOnModalVM)

                # Remove the modal component from the DOM by removing it in app.modals
                app.modals.remove(modalVM.modalRef)
            )

        onModalTemplateLoaded = () =>
            modalVM.modalDiv = $('#'+modalVM.modalDivId)
            modalVM.modalDiv.on('shown.bs.modal', (e) ->
                # Clear the binding - not having this causes madness - Topper
                $( modalVM.modalDiv ).unbind( 'shown.bs.modal' )
                # If there is an OnShow in the new modal, invoke it
                if modalVM.OnShow then modalVM.OnShow()
            )
            if not app.currentModalVM? # If no existing stack, just show the view model
                ShowModal(modalVM)
            else# Existing stack, need to hide existing modal and show ours

                # Here we drop the existing OnHide (hidden.bs.modal) handler,
                # and replace it with one that will have no animation..
                # Then once we've hidden the modal and shown the new one
                # we re-attach the original OnHide function - which is important

                # Clear the existing binding - not having this causes madness - Topper
                $( app.currentModalVM.modalDiv ).unbind( 'hidden.bs.modal' )

                # Create the new onHide ('hidden.bs.modal') handler
                app.currentModalVM.modalDiv.on('hidden.bs.modal', (e) ->
                    # Clear the binding - not having this causes madness - Topper
                    $( app.currentModalVM.modalDiv ).unbind( 'hidden.bs.modal' )
                    # Tell the new modal that it is stacked on another so we can restore
                    modalVM.stackedOnModalVM = app.currentModalVM
                    # Turn off the fade-in effect
                    modalVM.allowFade(false)
                    # Show the new modal
                    ShowModal(modalVM)
                )
                # Turn off the fade-out effect
                app.currentModalVM.allowFade(false)
                # Hide the current modal
                app.currentModalVM.modalDiv.modal("hide")


        subscription = modalVM.templateLoaded.subscribe (isLoaded) ->
            if isLoaded
                onModalTemplateLoaded()
                # Remove this subscription
                subscription.dispose()
                subscription = undefined
        return

    ohlife2::CloseModal = (closingModalVM) ->
        # Close it
        closingModalVM.modalDiv.modal("hide")
        return



    # Generic OnShow set-up - wait's until a component is loaded to call OnShow
    # Requires that the component has the following line in it template footer:
    # <!-- IMPORTANT - Mark this template loaded -->
    # <div data-bind="text:templateLoaded(true)" style="display:none"></div>
    ohlife2::InitOnShow = (VM) ->
        if not VM.isLoaded?
            #console.error("You need an 'isLoaded' variable if you are using InitOnShow()")
            return
        VM.templateLoaded = app.ko.observable(false)
        # Watch for changes to templateLoaded
        subscription = VM.templateLoaded.subscribe (isLoaded) ->
            if isLoaded
                if VM.OnShow then VM.OnShow()
                subscription.dispose()
                subscription = undefined
        return

    ohlife2::setupRedactorPlugins = () ->
        window.RedactorPlugins = {}  unless window.RedactorPlugins
        # insertSigVar plugin
        if not window.RedactorPlugins.insertSigVar?
            window.RedactorPlugins.insertSigVar = ->
                init: ->
                    @insertSigVarOpts = [
                        ["Mailbox - Email", "{%mailbox.email%}"]
                        ["Mailbox - Name", "{%mailbox.name%}"]
                        ["User - Full Name", "{%user.fullName%}"]
                        ["User - First Name", "{%user.firstName%}"]
                        ["User - Email", "{%user.email%}"]
                        ["User - Phone", "{%user.phone%}"]
                        ["User - Job Title", "{%user.jobTitle,fallback=%}"]
                    ]
                    dropOpts = {}
                    for opt, i in @insertSigVarOpts
                        dropOpts[i] = { title:opt[0].replace('-','&mdash;'), func:@insertSigVar.insertSigVarCB }
                    button = @button.add("insertVar", "Insert Variable")
                    @button.addDropdown button, dropOpts
                    return
                insertSigVarCB: (iStr) ->
                    i = parseInt(iStr,10)
                    html = @insertSigVarOpts[i][1]
                    @insert.html html
                    return
        # insertReplyVar plugin
        if not window.RedactorPlugins.insertReplyVar?
            window.RedactorPlugins.insertReplyVar = ->
                init: ->
                    @insertReplyVarOpts = [
                        ["Mailbox - Email", "{%mailbox.email%}"]
                        ["Mailbox - Name", "{%mailbox.name%}"]
                        ["Ticket - Id","{%ticket.id%}"]
                        ["Customer - Full Name","{%customer.fullName%}"]
                        ["Customer - First Name","{%customer.firstName%}"]
                        ["Customer - Email","{%customer.email%}"]
                        ["User - Full Name", "{%user.fullName%}"]
                        ["User - First Name", "{%user.firstName%}"]
                        ["User - Email", "{%user.email%}"]
                        ["User - Phone", "{%user.phone%}"]
                        ["User - Job Title", "{%user.jobTitle,fallback=%}"]
                    ]
                    dropOpts = {}
                    for opt, i in @insertReplyVarOpts
                        dropOpts[i] = { title:opt[0].replace('-','&mdash;'), func:@insertReplyVar.insertReplyVarCB }
                    button = @button.add("insertVar", "Insert Variable")
                    @button.addDropdown button, dropOpts
                    return
                insertReplyVarCB: (iStr) ->
                    i = parseInt(iStr,10)
                    html = @insertReplyVarOpts[i][1]
                    @insert.html html
                    return
        # insertSavedReply plugin
        if not window.RedactorPlugins.insertSavedReply?
            window.RedactorPlugins.insertSavedReply = ->
                init: ->
                    if window.RedactorPlugins.insertSavedReply.inboxId is undefined then return
                    $.getJSON( '/v1/inboxes/' + window.RedactorPlugins.insertSavedReply.inboxId + '/savedreplies.json', (response,d,xhr) =>
                        if not response.replies? or response.replies.length is 0 then return
                        @savedReplies = response.replies
                        dropOpts = {}
                        for opt, i in @savedReplies
                            dropOpts[i] = { title:opt.name, func:@insertSavedReply.insertSavedReplyCB }
                        button = @button.add("insertSavedReply", "Insert Save Reply")
                        @button.addDropdown button, dropOpts
                    )

                    return
                insertSavedReplyCB: (iStr) ->
                    i = parseInt(iStr,10)
                    html = @savedReplies[i].reply
                    @insert.html html
                    return



    ###
    Generic Utility functions
    ###
    ohlife2::IsEmail = (email) ->
        regex = /^([a-zA-Z0-9_.+-])+\@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$/
        return regex.test(email)



    ohlife2::convertZuluStringToMoment = (dateStr) ->
        return moment.min( moment( dateStr, 'YYYY-MM-DDTHH:mm:ssZ'), moment())


    ohlife2::SafeHTML = (val) ->
        return String(val).replace(/&/g,"&amp;").replace(/\"/g,"&quot;").replace(/\'/g,"&#39;").replace(/</g,"&lt;").replace(/>/g,"&gt;")

    ohlife2::URLSafeString = (val) ->
        return String(val).replace(/\s/g, '-').replace(/[^a-zA-Z0-9-_]/g, '').toLowerCase()


    ###
    Given a file size, returns a human friendly text string such as "27KB" for "3.2 Mb (medium)"
    use like tw.getFriendlyBytes( 888, {getHint:true} );
    @method getFriendlyBytes
    @param {Number} The size of a file.
    ###
    ohlife2::GetFriendlyBytes = (size, opts) ->
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
    ohlife2::NumberFormat = (num) ->
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",")


    ohlife2::GetAttachmentIcon = (filename, mimetype) ->
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
            psd: ["psd"]
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
            ps: ["ps"]
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
    These are helper functions that any page view can use to ensure that
    part of the route to the page is always valid.
    It can also be used to rember the last valid route (dashboard uses this)

    This was also required because of the way that current page route (url)
    is now mapped to the app.route{} and each part is an observable variable,
    causeds components to not reload... so defaults couldn't be easily set-up.

    There were other ways to solve the problem but I wanted it all oncapsulated
    away neatly.

    I know this looks complex but when I wrote the code I realized that this
    could be make more generic and used in several places - it's the generic
    nature of it that makes it look complex.

    options (opts):
        vm: Required. The view model (page) to use this on
        routePart: Required. The part of the route that we are watching, an observable inapp.route[routePart]
        validOpts: Required. A list of valid values for the route on this view model (page)
        rememberIn: Option. "lastDashboardPageTab"}
    ###
    ohlife2::ensureValidRoute = (routePart,opts) ->
        vm = opts.vm

        # We only want to enforce the validRoute if the part of the route matches
        # This prevents subscriptions that are not removed quickly enough causing problems
        if opts.onlyIfSame?
            if not opts.onlyIfSameInitialVal?
                opts.onlyIfSameInitialVal = app.currentRoute()[opts.onlyIfSame]
            else
                if app.currentRoute()[opts.onlyIfSame] != opts.onlyIfSameInitialVal
                    return
        if not opts.onlyIfSame?
            console.info("Dev note: Missing 'onlyIfSame'", routePart, opts)

        ##############################################################################
        # PART 1 - Setup watch on the current route
        ##############################################################################
        if not vm.validRouteSubscription?
            vm.validRouteSubscriptionsList = {}

            # Setup the method to remove this subscription
            subscriptionFn = (vm) ->
                @vm = vm
                return (newRouteVal) ->
                    routeObj = app.currentRoute()
                    for routePart, opts of vm.validRouteSubscriptionsList
                        app.ensureValidRoute(routePart, opts)
                    return

            # Create the app.CurrentRoute subscripton
            vm.validRouteSubscription = app.currentRoute.subscribe( subscriptionFn(vm) )


        # Add these opts to the list of routes to keep an eye on
        if not vm.validRouteSubscriptionsList[routePart]?
            vm.validRouteSubscriptionsList[routePart] = opts

        ##############################################################################
        # PART 2 - Check to see if this latest version of the route value is ok
        ##############################################################################
        routeObj = app.currentRoute()
        routePartVal = routeObj[routePart]

        # if old val is ok, just remember and exit
        if typeof routePartVal isnt 'undefined' and opts.validOpts.indexOf(routePartVal) >= 0
            if opts.rememberIn? then lscache.set( opts.rememberIn, routePartVal )
            return

        ##############################################################################
        # Part 3 - We have received an invalid value - pick a good one instead
        ##############################################################################

        # If we don't have a good value already, try to get the last one used
        if opts.rememberIn?
            newVal = lscache.get(opts.rememberIn)

        # If the route is still bad, then use the first available one
        if opts.validOpts.indexOf(newVal) is -1
            newVal = opts.validOpts[0]

        # Update the route with the new valid value
        routeObj[routePart] = newVal

        # Save the new value
        if opts.rememberIn? then lscache.set( opts.rememberIn, newVal )

        return

    # Helper function to dispose the subscriptions mapped by the above function
    ohlife2::removeValidRouteSubscriptions = (vm) ->
        if vm.validRouteSubscription?
            vm.validRouteSubscription.dispose()



    # Helper function to dispose the subscriptions mapped by the above function
    ohlife2::SaveLoggedInUser = (vm) ->
        lscache.set "loggedInUser", mapping.toJS( @loggedInUser )

    ohlife2::ClearLoggedInUser = (vm) ->
        lscache.set "loggedInUser", null
        @loggedInUser = null



    ###
    Takes care of the need for either a ? or & when extending a url
    use like url = extendURL(url,'key=val');
    @method ExtendURL
    @param {String} The url to extend.
    @param {Object} Param pair. The bit to append to the url. should look like 'x=y'}
    ###
    ohlife2::ExtendURL = (url,str) ->
        url += if url.indexOf("?")==-1 then '?' else '&'
        if typeof str is "string"
            url += str
        else if typeof str is "object"
            params = $.param( str )
            url += params
        return url




    return ohlife2