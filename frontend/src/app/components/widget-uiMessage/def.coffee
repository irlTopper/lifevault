define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        @watch = params.watch
        @watch.subscribe (newValue) =>
            if newValue? and typeof newValue.timer == "number"
                @timer = setTimeout( =>
                    @dismiss()
                , newValue.timer )

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::dismiss = (o,event) ->
        if @timer?
            clearTimeout( @timer )
            @timer = null

        # Fade effect on updated columns
        $("#uiMessage").fadeOut({
            complete: () =>
                @watch(null)
        })

        return


    return {
        viewModel: VM
        template: templateMarkup
    }