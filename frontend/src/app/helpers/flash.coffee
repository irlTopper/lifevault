define [], ->

    Flash = (app) ->
        @app = app
        return

    # Will display a message to the user - a type can be passed in options
    # valid "type"'s' are 'error', 'warning', 'success' and 'info'
    Flash::Info = (msg, inOpts) ->
        opts = {
            text: msg,
            type: "info",
            timer: 3000
        }
        if inOpts?
            $.extend(opts, inOpts)

        @app.uiMessage(opts)
        return

    # Cnvenience wrapper for FlashMessage with type set to error
    Flash::Success = (msg, inOpts) ->
        opts = {
            type: "success"
            timer: 800
        }
        if inOpts?
            $.extend(opts, inOpts)
        @Info(msg, opts)
        return

    # Cnvenience wrapper for FlashMessage with type set to error
    Flash::Error = (msg, inOpts) ->
        opts = {
            type: "error"
            timer: 4000
        }
        if inOpts?
            $.extend(opts, inOpts)

        @Info(msg, opts)
        return

    return Flash