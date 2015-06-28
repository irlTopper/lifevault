define [
    "knockout"
    "text!./template.html"
    "jquery.validate"
], (ko, templateMarkup, jqueryvalidate) ->
    VM = (params) ->
        # Handle params
        @title = params.title
        @question = params.question
        @callback = params.callback
        @buttonText = if params.buttonText? then params.buttonText else "Save"

        # Vars
        @answer = ko.observable("")

        # Setup OnShow
        app.modal.Init('prompt', @, params)
        return

    VM::OnShow = () ->
        # Focus first field
        @formId = @modalDivId+'-form'
        $('#'+@formId).validate({onsubmit: false})
        @modalDiv.find("input[type=text]").first().focus()
        return

    VM::Close = () ->
        app.modal.Close(@, false)
        return


    VM::Cancel = ->
        @Close()
        return

    VM::Finished = ->
        form = $('#'+@formId)
        if not form.valid()
            return false

        @Close()
        if @callback?
            @callback( @answer() )

    return {
        viewModel: VM
        template: templateMarkup
    }