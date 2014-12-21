define ["knockout", "crossroads", "hasher", "lscache"], (ko, crossroads, hasher, lscache) ->

    Router = (config) ->
        crossroads.shouldTypecast = true # default = false
        currentRoute = this.currentRoute = ko.observable({})

        ko.utils.arrayForEach config.routes, (route) ->

            # default params for this app
            if not route.params.isLoggedInPage?
                route.params.isLoggedInPage = true

            crossroadsRoute = crossroads.addRoute route.url, (requestParams) ->

                # Extract ids from URLs like /tickets/34-Name here:
                if requestParams.inboxURLCode?
                    requestParams.inboxId = parseInt requestParams.inboxURLCode, 10

                if requestParams.helpsiteURLCode?
                    requestParams.siteId = parseInt requestParams.helpsiteURLCode, 10

                # Home page redirect
                if route.params.page is '-startPageRedirect-'
                    if typeof window.app is 'undefined' or typeof window.app.loggedInUser is 'undefined' or window.app.loggedInUser is null or window.app.loggedInUser.id() is 0
                        route.params.page = 'login'
                        route.params.isLoggedInPage = false
                    else
                        route.params.page = 'dashboard'
                        route.params.dashboardPageTab = ( lscache.get( "lastDashboardPageTab" ) || "inboxes" )
                        route.params.isLoggedInPage = true

                currentRoute ko.utils.extend( requestParams, route.params )
                return

            ###
                If we have an {id} part, we only want it to match numbers.
            ###
            if route.url.indexOf("{id}")>-1
                crossroadsRoute.rules = crossroadsRoute.rules || {}
                crossroadsRoute.rules.id = (value, request, valuesObj) ->
                    if isNaN(value) then return false
                    return true

            ###
                inboxURLCode must start with a number that identifies the inbox
                by id eg "45-Support"
            ###
            if route.url.indexOf(":inboxURLCode:") > -1
                crossroadsRoute.rules = crossroadsRoute.rules || {}
                crossroadsRoute.rules.inboxURLCode = (value, request, valuesObj) ->
                    inboxId = parseInt( value, 10 )
                    if isNaN( inboxId ) then return false
                    return true

            ###
                inboxView
            ###
            if route.url.indexOf(":inboxView:") > -1
                crossroadsRoute.rules = crossroadsRoute.rules || {}
                crossroadsRoute.rules.inboxView = (value, request, valuesObj) ->
                    value = value.toLowerCase()
                    if value == "newticket" then return false# We won't want to make this special case which has it's own route
                    return true


        @hasher = hasher# needs to be exposed for cool stuff

        return

    Router::init = ->
        parseHash = (newHash, oldHash) ->
            crossroads.parse(newHash)
        crossroads.normalizeFn = crossroads.NORM_AS_OBJECT
        hasher.initialized.add(parseHash)
        hasher.changed.add(parseHash)
        hasher.init()



    routes = [
            # main nav
            { url: 'dashboard/:dashboardPageTab:', params: { page: 'dashboard' } }
            { url: 'tickets/{id}', params: { page: 'tickets', pageMain: 'ticket' } }
            { url: 'unified/tickets/{id}', params: { page: 'tickets', pageMain: 'ticket', inUnified: true } }
            { url: 'tickets/:inboxURLCode:/newticket', params: { page: 'tickets', pageMain: 'ticket-new' } }
            { url: 'tickets/:inboxURLCode:/:inboxView:', params: { page: 'tickets', pageMain: 'inbox' } }

            { url: 'customers/{id}', params: { page: 'customer' } }
            { url: 'reports/:pageTab:', params: { page: 'reports' } }
            { url: 'helpdocs/articles/:helpdocId:', params: { page: 'helpdocs', pageMain: 'helpdocs-article' } }
            { url: 'helpdocs/:helpsiteURLCode:', params: { page: 'helpdocs', pageMain: 'helpdocs-list' } }
            { url: 'helpdocs/:helpsiteURLCode:/category/:categoryId:', params: { page: 'helpdocs', pageMain: 'helpdocs-list' } }
            { url: 'helpdocs/:helpsiteURLCode:/category/:categoryId:/new', params: { page: 'helpdocs', pageMain: 'helpdocs-article' } }
            { url: 'helpdocs/:helpsiteURLCode:/new', params: { page: 'helpdocs', pageMain: 'helpdocs-article' } }

            # extra pages
            { url: 'search/:term1:/:term2:/:term3:/:term4:', params: { page: 'search' } }
            { url: 'settings/:pageTab:/:view:/:view2:/:view3:/:view4:', params: { page: 'settings' } }
            { url: 'errorLoadingMsg', params: { page: 'errorLoadingMsg' } }
            # My pages
            { url: 'myprofile/:view:', params: { page: 'myprofile' } }
            { url: 'plan', params: { page: 'plan' } }
            # Logged out pages
            { url: 'login', params: { page: 'login', isLoggedInPage:false } }
            { url: 'resetpassword/:userId:/:name:/:token:', params: { page: 'reset-password', isLoggedInPage:false } }
            { url: 'forgotpassword/:username:', params: { page: 'forgot-password', isLoggedInPage:false } }
            { url: 'happiness-rating/:ticketId:/vote/:token:/:rating:', params: { page: 'happiness-rating', isLoggedInPage:false } }
            # default page
            { url: '/:startPage:', params: { page: '-startPageRedirect-' } }

        ]

    return new Router({routes:routes})