define [
    "knockout"
    "text!./template.html"
], (ko, template) ->
    VM = (params) ->
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::refreshIfNeeded = ->
        # Reload the main page component
        if app.hasher.getHash() is "dashboard/inboxes"
            app.RefreshMainPage()
            return true

        return true

    VM::ShowOrHide = (e, o) ->
        if app.currentRoute().page is "tickets"
            btn = $('#menuTicket')

            dataToggle = $(btn).data('toggle-target')
            dataParentToggle = $(btn).data('toggle-parent')

            $(dataToggle).toggleClass('menu-open')
            $(dataParentToggle).toggleClass('menu-open')

            $('#ticket-main-content').toggleClass('menu-open')
            $('#new-ticket-content').toggleClass('menu-open')
        else
            return true

    VM::OnClickReports = () ->
        if not app.loggedInUser.onboarding.exploreReports()
            app.UpdateOnboarding 'exploreReports', true

        app.GoTo 'reports/overview'
        return

    VM::OnClickHelpdocs = () ->
        if not app.loggedInUser.onboarding.exploreHelpsites()
            app.UpdateOnboarding 'exploreHelpsites', true

        app.GoTo 'helpdocs'
        return

    return {
        viewModel: VM
        template: template
    }
