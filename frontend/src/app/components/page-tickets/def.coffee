###
This is basically just a holder that loads in the left hand side and right hand side
for all ticket/inbox pages
###
define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->

    VM = (params) ->
        app.isLoading true
        @readyToShow = app.ko.observable false
        @ticket = app.ko.observable()
        @pageMainComponent = app.ko.observable()

        # Figure out which component we want to render and update it when the page changes
        @updatePageMainComponent()
        @routeSub = app.currentRoute.subscribe () =>
            @updatePageMainComponent()

        app.mainHolderClass 'wrap-inbox'

        # Mark the page as loaded
        app.isLoading(false)
        @readyToShow true

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->
        $(".custom_scroll").mCustomScrollbar({theme: 'minimal-dark', scrollInertia: 0})
        return

    # Remember which component we want to render and update it when the page changes
    # This technique prevents flicker by reusing a component but has a catch, when the
    # user navigates away, we get a new component which we try to render.
    # We limit this problem by ignoring everything except "inbox" and "ticket"
    VM::updatePageMainComponent = () ->
        if ["inbox","ticket","ticket-new"].indexOf(app.currentRoute().pageMain) is -1 then return
        @pageMainComponent("pageMain-" + app.currentRoute().pageMain)
        return

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->
        @routeSub.dispose()
        return

    return {
        viewModel: VM
        template: templateMarkup
    }