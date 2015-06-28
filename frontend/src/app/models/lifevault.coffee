define [
    'jquery'
    'jquery-ui'

    'lscache'
    'router'
    'app/helpers/utility'
    'app/helpers/flash'
    'app/helpers/error'
    'app/helpers/modal'
    'app/helpers/editor'
    'app/helpers/keybindings'
    'app/helpers/redactor'
    'app/helpers/helpdoc'
    'knockout'
    'knockout-mapping'
    'punches'
    'app/socket'
    'userModel'
    'mousetrap'
    'lodash'
    'config-components'
    'app/helpers/mail'
    'jquery-mockjax'
    # Unnamed omcponents after this
    'knockout-switch-case'
    'knockout-repeat'
    'bootstrap-select'
    'mCustomScrollbar'
    'bindings-ladda'
    'bindings-misc'
    'mousetrap-global'
    'raven'
    'jquery.validate',
    'knockout-projections',
    'redactor',
    'bootstrap',
    'modernizr'
    'highcharts'
],
(jQuery, jui, lscache, router, Utility, FlashHelper, ErrorHelper, ModalHelper, RedactorPlugins, KeyBindings, RedactorHelper, HelpdocHelper, ko, mapping, punches, Socket, userModel, Mousetrap, _, components, MailHelper, mockjax) ->

    LifeVault = ->
        @initMockjax()

        if not window.twDeskConfig?
            window.twDeskConfig =
                "isDeveloperMode": false

        # Production mode redirects
        if not twDeskConfig.isDeveloperMode
            # ensure HTTPS when in production mode
            if window.location.protocol != "https:" and (window.location.href.indexOf("teamwork.com") isnt -1 or window.location.href.indexOf("teamworkpm.net") isnt -1)
                window.location.href = "https:" + window.location.href.substring(window.location.protocol.length)
                return

            # Deprecating lifevault.com, move them to teamwork.com/desk
            if window.location.host.indexOf("lifevault.com") isnt -1
                window.location.href = window.location.href.replace 'lifevault.com', 'teamwork.com/desk'
                return

        @hasLocalStorageSupport = true# See startup.coffee - this is needed
        @ko = ko
        @mapping = mapping
        @lscache = lscache
        @utility = new Utility()
        @socket = new Socket(@)
        @flash = new FlashHelper(@)
        @error = new ErrorHelper(@)
        @modal = new ModalHelper(@)
        @redactor = new RedactorHelper()
        @keybindings = new KeyBindings()
        @helpdoc = new HelpdocHelper()
        @mailHelper = new MailHelper()
        @essentialDataIsLoaded = ko.observable(false)
        @mainHolderClass = ko.observable('')
        @isLoading = ko.observable(false)
        @pageLoadingErrorXHR = ko.observable(null)
        @pageComponentToShow = ko.observable('')
        @helpDocsSites = ko.observableArray []
        @router = router
        @hasher = router.hasher
        @currentRoute = router.currentRoute
        @users = ko.observableArray()
        @currentRoute.subscribe (newRoute) => @onUpdateRoute(newRoute)
        @uiMessage = ko.observable null #used for displaying a message to the user
        @modals = ko.observableArray() #used or tracking loaded modals components
        @dates =  ko.observableArray()

        # Prevent ajax caching in old browsers (Internet Explorer is retarded)
        # "nocache" header added to server side should fix IE10+
        # This is for <= IE9 - must do client side
        # If we have problems wih IE11 later, look at this:
        # http://stackoverflow.com/questions/19999388/jquery-check-if-user-is-using-ie
        if window.navigator.userAgent.indexOf("MSIE ") > -1
            $.ajaxSetup({ cache: false })

        # Append version number to all ajax requests - handy for logging later

        # Remember if the app is running on a mobile device
        @isMobile = ((a) ->
            if /android.+mobile|avantgo|bada\/|blackberry|blazer|ipad|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|symbian|treo|up\.(browser|link)|vodafone|wap|windows (ce|phone)|xda|xiino/i.test(a) or /1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|e\-|e\/|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(di|rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|xda(\-|2|g)|yas\-|your|zeto|zte\-/i.test(a.substr(0, 4))
                true
            else
                false
        )(navigator.userAgent or navigator.vendor or window.opera)
        @isIOS = navigator.userAgent.match(/(iPad|iPhone|iPod)/g)

        ua = navigator.userAgent.toLowerCase()
        @isWindowsSafari = (ua.indexOf("safari/") isnt -1 and ua.indexOf("windows") isnt -1 and ua.indexOf("chrom") is -1)

        # Redactor plugin drop downs - save us setting it up all the time
        @redactorPlugings = new RedactorPlugins()

        # We need to set this so we're telling the server that this is
        # the LifeVault app making the requests (prevents auth popups)
        jQuery.ajaxSetup({
            headers: {
                'lv': '0.1'
            }
        })

        # We use the lscache to remember if we were logged in before this reload
        # This is secure because the server side is really in charge and the first
        # ajax load will fail if the user isn't logged in.
        # The purpose of this is basically to save us having to do GET /login.json
        # whenever the user does a full reload. ps. If this fails, it's null = not logged in.
        @LoadUser()

        # Start the application by initializing the router and root binding
        # Done before we start really using the app.Init() to ensure there are no errors
        @initSafeConsole()

        return @

    LifeVault::initMockjax = () ->
        console.info "initMockjax"
        $.mockjax({
            url: "v1/journal/dates.json",
            responseText: {
                status: "success",
                dates: [
                    "2015-04-02"
                    "2015-04-01"
                    "2015-03-26"
                ]
            }
        })
        $.mockjax({
            url: "v1/settings.json",
            responseText: {
                status: "success",
                fortune: "Are you a mock turtle?"
            }
        })
        console.info "initMockjax done"
        return



    # Init - put anything in here that requires looking up settings or
    # requires window.app to be defined globally (like socket.Start)
    LifeVault::Init = (opts) ->
        window.app = @

        # This displays the app - we need to do this asap so put anything not essential after this
        ko.applyBindings {}

        # Start the socket server if user is logged in
        if @loggedInUser? then @socket.Start()

         # Pretend the navigation just changed to kick things off
        @onUpdateRoute(@currentRoute())

        @router.init()

        @initRaven() # Raven is used for tracking javascript errors
        @initKeyboardShortcuts() # Kayboard shortcuts
        @initBrowserSync() # Browser sync reloads automcatically in dev mode

        # Preload some extra CSS and javascript
        @preloadCSS()
        return

    LifeVault::initBrowserSync = () ->
        if not twDeskConfig.isDeveloperMode then return
        scriptPath = "//" + location.hostname + ":50243/browser-sync-client.1.3.7.js"
        jQuery.getScript(scriptPath)
        return

    LifeVault::LoadUser = () ->
        $.ajax({
            url: 'v1/me.json'
            success: (response, d, xhr) =>
                @handleUserResponse response
        }).error((response, d, xhr) =>
            @loggedInUser = null
        ).always () =>
            @Init()

    LifeVault::UpdateOnboarding = (flag, set) ->
        $.ajax({
            url: 'v1/me/onboarding.json'
            type: 'PUT'
            data:
                flag: flag
                set: set
            success: (response, d, xhr) =>
                @handleUserResponse response
        }).error((response, d, xhr) =>
            app.error.Ajax xhr
        )

    # Utility function for handling user responses for the
    # currently logged in user
    LifeVault::handleUserResponse = (response) ->
        @loggedInUser = mapping.fromJS(response, userModel.mapping).user
        console.info "userModel", userModel.mapping
        return

    # Here we load CSS that aren't essential for the initial load but needed
    LifeVault::preloadCSS = () ->
        h = $('head')
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/bower/Ladda/dist/ladda-themeless.min.css') )
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/bower/jquery-minicolors/jquery.minicolors.css') )
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/redactor/redactor.css') )
        h.append( $('<link rel="stylesheet" type="text/css" />').attr('href', 'libs/bower/ko-calendar/dist/ko-calendar.min.css') )
        return

    LifeVault::initKeyboardShortcuts = ->
        Mousetrap.bind "?", =>
            app.modal.Show("keyboard-shortcuts", {}, @)
            return

        @shiftKeyDown = false
        Mousetrap.bind "shift", =>
            @shiftKeyDown = true
            return
        , "keydown"
        Mousetrap.bind "shift", =>
            @shiftKeyDown = false
            return
        , "keyup"

        @commandKeyDown = false
        Mousetrap.bind "command", =>
            @commandKeyDown = true
            return
        , "keydown"
        Mousetrap.bind "command", =>
            @commandKeyDown = false
            return
        , "keyup"

        @ctrlKeyDown = false
        Mousetrap.bind "ctrl", =>
            @ctrlKeyDown = true
            return
        , "keydown"
        Mousetrap.bind "ctrl", =>
            @ctrlKeyDown = false
            return
        , "keyup"

        return

    LifeVault::UpdateDates = (inOpts) ->
        return $.getJSON('v1/journal/dates.json', (response,d,xhr) =>
            @dates( mapping.fromJS(response).dates )
        ).error (response,d,xhr) =>
            # isAutoUpdate = silent error
            if opts? && opts.isAutoUpdate? && opts.isAutoUpdate then return
            return @redirectToLogin()

    LifeVault::UpdateSettings = (opts) ->
        return $.getJSON( 'v1/settings.json', (response,d,xhr) =>
            if @settingsResponse?
                mapping.fromJS response, @settingsResponse
            else
                @settingsResponse = mapping.fromJS response
                @settings = @settingsResponse.settings
        ).error (response,d,xhr) =>
            # isAutoUpdate = silent error
            if opts? && opts.isAutoUpdate? && opts.isAutoUpdate then return
            return @redirectToLogin()


    LifeVault::loadEssentialData = ( callback ) ->
        ## This awkward looking syntax loads the json in parallel and only when it's finished
        ## toggles essentialDataIsLoaded to true.
        ## Todo: Handle network errors elegantly, maybe switch to the network error page
        $.when(
            @UpdateDates { forceUpdate: true }
        ,
            @UpdateSettings()
        ).then =>
            @essentialDataIsLoaded true
            # Perform the callback - this will show the requested logged in page
            callback()

        return

    LifeVault::redirectToLogin = ( callback ) ->
         # Redirect to login but remember this page
        if app.currentRoute().page isnt "login"
            app.requestedRouteBeforeLoginRedirect = app.currentRoute()
            app.GoTo "login"
        return false

    # This will ensure that console messages don't blow up in old browsers
    LifeVault::initSafeConsole = ->
        if window.console is undefined
            window.console =
                assert: -> return
                log: -> return
                warn: -> return
                error: -> return
                debug: -> return
                dir: -> return
                info: -> return
                clear: -> return
        return

    LifeVault::initRaven = ->
        return

    # For navigating around the app
    # Pass the silent flag to change it without the app updating
    # Handy for when something is added.
    LifeVault::GoTo = (hash,silent=false) ->
        console.assert( hash.indexOf("http") is -1, "Don't use GoTo for full URLs" )

        hash = hash.toLowerCase()

        if not silent
            @hasher.setHash(hash)
        else
            @hasher.changed.active = false
            @hasher.setHash(hash)
            @hasher.changed.active = true
        return

    LifeVault::onUpdateRoute = (newRoute) ->
        # Load the essential data asap - logged in pages won't show until this is done
        if newRoute.isLoggedInPage and not @essentialDataIsLoaded()
            if not @loggedInUser?
                @redirectToLogin()
                return
            @pageComponentToShow( 'page-loading' )
            @loadEssentialData () =>
                @pageComponentToShow( 'page-' + newRoute.page )
        else
            @pageComponentToShow( 'page-' + newRoute.page )

        # We store a userPageStatus for each route, always starts with "viewing"
        @currentRoute().userPageStatus = "viewing"

        if newRoute.isLoggedInPage && @socket
            @socket.NotifyNewRoute()

        return


    # RefreshMainPage re-renders the main page component
    # This works by by invalidating the "pageComponentToShow"
    # used to decide which main page component to display.
    LifeVault::RefreshMainPage = () ->
        app.pageComponentToShow.valueHasMutated()
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
    LifeVault::MissingEssentialData = () ->
        # If we've tried the reload already and still no luck, then go to the dash
        if lscache.get("MissingEssentialData-tryingReload") isnt null
            lscache.set "MissingEssentialData-tryingReload", null
            app.GoTo "dashboard"
            app.flash.Error("We couldn't load the page you requested &mdash; you might be denied access or it might have been deleted.")
            return

        lscache.set "MissingEssentialData-tryingReload", true, 1000
        location.reload()
        return



    # Use this is you were trying to load essential data for a page and there was a problem,
    # We will show an generic error message page and provide a try again button.
    LifeVault::showErrorLoadingMsg = (xhr) ->
        @pageLoadingErrorXHR(xhr)
        return



    # Generic OnShow set-up - wait's until a component is loaded to call OnShow
    # Requires that the component has the following line in it template footer:
    # <!-- IMPORTANT - Mark this template loaded -->
    # <div data-bind="template:{afterRender:function(){templateLoaded(true)}}" class="hidden"></div>
    LifeVault::InitOnShow = (VM,readyToShow=true) ->
        if not VM.readyToShow?
            VM.readyToShow = ko.observable(readyToShow)
        VM.templateLoaded = app.ko.observable(false)
        # Watch for templateLoaded to be set to true, then call OnShow()
        subscription = VM.templateLoaded.subscribe (isTemplateLoaded) ->
            if isTemplateLoaded
                if VM.OnShow then VM.OnShow()
                $('.pops').popover({ trigger:"hover", html:true })
                subscription.dispose()
                subscription = undefined
        return



    ###
    These are helper functions that any page view can use to ensure that
    part of the route to the page is always valid.
    It can also be used to rember the last valid route (dashboard uses this)

    This was also required because of the way that current page route (url)
    is now mapped to the app.route{} and each part is an observable variable,
    causes components to not reload... so defaults couldn't be easily set-up.

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
    LifeVault::ensureValidRoute = (routePart,opts) ->
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
            console.log("Dev note: Missing 'onlyIfSame'", routePart, opts)

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
    LifeVault::removeValidRouteSubscriptions = (vm) ->
        if vm.validRouteSubscription?
            vm.validRouteSubscription.dispose()

    # Helper function to dispose the subscriptions mapped by the above function
    LifeVault::SaveLoggedInUser = (vm) ->
        lscache.set "loggedInUser", mapping.toJS( @loggedInUser )

    LifeVault::Logout = () ->
        lscache.set "loggedInUser", null
        @loggedInUser = null
        app.socket.Stop()
        app.GoTo 'login'
        $.getJSON( 'v1/logout.json' )



    return LifeVault