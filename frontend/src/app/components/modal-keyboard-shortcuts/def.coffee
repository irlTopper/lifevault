define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->

    VM = (params) ->
        @title = "Keyboard Shortcuts"

        # Setup OnShow
        app.modal.Init('keyboardShortcuts', @, params)
        return

    VM::Close = ->
        app.modal.Close(@, false)
        return

    return {
        viewModel: VM
        template: templateMarkup
    }