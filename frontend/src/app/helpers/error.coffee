define ['app/helpers/flash'], (FlashHelper) ->

    Error = (app) ->
        @app = app
        @flash = new FlashHelper(@app)

        # Some messages to save us having them everywhere in code
        @messages =
            SessionExpired: "Sorry it looks like your session has expired.",
            AccessDenied: "The server has denied you access. Your permissions may have been changed. If this continues try logging out and back in again.",
            NetworkTimeout: "<strong>Network timeout</strong> &mdash; Please check your internet connection and try again. Email desk@teamwork.com if you still have problems.",
            ServerRejected: "<strong>Server rejected</strong> &mdash; Sorry the server rejected this request.",
            ServerRejectedWithMessage: "The server rejected this request with the message:",
            ServerError: "<strong>Server error</strong> &mdash; Sorry the server ran into a problem processing this request."

        return

    # Use this if there was an error performing an action
    Error::Ajax = (xhr) ->
        explOpts = ["Whoops!","Yikes!","Blimey!","Oh no!","Oh Dear!"]
        expl = explOpts[Math.floor(explOpts.length*Math.random())]
        if xhr.status == 401 # Expired session
            @flash.Error( @messages.SessionExpired + " &ndash; <a href=\"javascript:document.location.reload()\">Login again to continue</a>" )
        else if xhr.status == 403 # Access Denied
            @flash.Error( @messages.AccessDenied )
        else if xhr.status == 400 # Server rejected
            if xhr.responseJSON? and xhr.responseJSON.message
                @flash.Error( "<strong>"+expl+"</strong> " + xhr.responseJSON.message )
            else if xhr.responseJSON? and xhr.responseJSON.errors?
                ers = _.values(xhr.responseJSON.errors)
                if ers[0].Message?
                    @flash.Error( "<strong>"+expl+"</strong> " + ers[0].Message )
                else
                    @flash.Error( "<strong>"+expl+"</strong> " + ers[0] )
            else
                @flash.Error( @messages.ServerRejected )
        else if xhr.status == 0 # Network timeout
            @flash.Error( @messages.NetworkTimeout )
        else
            if xhr.responseJSON? and xhr.responseJSON.message
                @flash.Error( "<strong>"+expl+"</strong> &mdash; " + xhr.responseJSON.message )
            else if xhr.responseJSON? and xhr.responseJSON.errors
                @flash.Error( "<strong>"+expl+"</strong> &mdash; " + xhr.responseJSON.errors[0] )
            else
                @flash.Error( @messages.ServerError )

        return

    return Error