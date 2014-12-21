define [
    "knockout"
    "text!./template.html"
], (ko, template) ->
    VM = (params) ->
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::ShowOrHide = (e, o) ->
        if app.currentRoute().page is "tickets"
            btn = $('#menuTicket')

            dataToggle = $(btn).data('toggle-target')
            dataParentToggle = $(btn).data('toggle-parent')

            $(dataToggle).toggleClass('menu-open')
            $(dataParentToggle).toggleClass('menu-open')

            $('#ticket-main-content').toggleClass('menu-open')
        else
            return true

    return {
        viewModel: VM
        template: template
    }
