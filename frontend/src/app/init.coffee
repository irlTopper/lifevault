# So the idea here is that we want only the absolute minimum needed to run here.
define [
    'knockout'
], (ko) ->

    @hasLocalStorageSupport = typeof(Storage) isnt "undefined"


    if typeof(Storage) is "undefined"
        window.app = @
        @pageComponentToShow = "page-login"
        @_template = "normal"
        @hasLocalStorageSupport = false

        # This displays the app - we need to do this asap so put anything not essential after this
        ko.applyBindings {}

    else
        require ["lifevault"], (app) ->
            ## Bada boom - create & start the application, assign to window for easy access to "app" from anywhere
            ko.punches.enableAll()
            window.app = new app()