define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        # Handle Params
        @title = params.title
        @question = params.question
        @callback = params.callback

        # Setup OnShow
        app.modal.Init('Confirm', @, params)
        return

    VM::OnShow = () ->
        $("#modalConfirmOKBut").focus()
        return

    VM::Close = () ->
        app.modal.Close @, false
        return


    VM::Cancel = ->
        @Close()

    VM::OK = ->
        @Close()

        if @callback?
            @callback()

    return {
        viewModel: VM
        template: templateMarkup
    }