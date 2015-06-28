define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->


    ###  ____       _
        / ___|  ___| |_ _   _ _ __
        \___ \ / _ \ __| | | | '_ \
         ___) |  __/ |_| |_| | |_) |
        |____/ \___|\__|\__,_| .__/
                             |_|
    # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # ###

    VM = (params) ->
        @title = params.title
        @text = params.text
        @icon = params.icon
        @onClick = params.onClick
        @clickText = params.clickText
        @parentVM = params.parentVM # optional
        app.InitOnShow(@)# Sets-up the OnShow() function
        return


    VM::OnShow = () ->
        if @parentVM.NotifyEvent?
            @parentVM.NotifyEvent({itemType:'component', actionType:'OnShow', vm:@})
        return







    ###  _____                 _     _   _                 _ _
        | ____|_   _____ _ __ | |_  | | | | __ _ _ __   __| | | ___ _ __ ___
        |  _| \ \ / / _ \ '_ \| __| | |_| |/ _` | '_ \ / _` | |/ _ \ '__/ __|
        | |___ \ V /  __/ | | | |_  |  _  | (_| | | | | (_| | |  __/ |  \__ \ - On...
        |_____| \_/ \___|_| |_|\__| |_| |_|\__,_|_| |_|\__,_|_|\___|_|  |___/
    # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # ###

    VM::ShowAddSite = () ->
        app.modal.Show "new-helpdocsite", { callback: app.RefreshMainPage }, @
        return




    ### _____
        |  _ \(_)___ _ __ | | __ _ _   _
        | | | | / __| '_ \| |/ _` | | | |
        | |_| | \__ \ |_) | | (_| | |_| |
        |____/|_|___/ .__/|_|\__,_|\__, |
                    |_|            |___/
    # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # ###

    return {
        viewModel: VM
        template: templateMarkup
    }
