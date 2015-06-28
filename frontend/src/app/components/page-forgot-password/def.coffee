define ['knockout', 'knockout-mapping', 'text!./template.html', 'userModel'], (ko, mapping, templateMarkup, userModel) ->

    VM = (params) ->
        @siteName = ko.observable ""
        @logo = ko.observable ""
        @isSubmitting = ko.observable false

        @email = ko.observable ""

        if app.currentRoute().username?
            @email app.currentRoute().username

        $.getJSON "v1/settings/branding.json", (response,d,xhr) =>
            @logo response.branding.logo

        .fail (xhr) ->
            return app.showErrorLoadingMsg(xhr)

        # Fix for background not showing
        $('body').attr('id', 'page-login')

        # Focus the password box
        $('#password').focus()

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::onSubmitForgotPassword = () ->
        @isSubmitting true

        $.ajax({
            url: "v1/users/forgotpassword.json",
            type: 'POST',
            dataType: 'json',
            data: {
                email: @email()
            }
            success: (response,d,xhr) =>
                app.flash.Success('A password reset was issued to your email address. Check your email!', { timer: 5000 })

                app.GoTo 'login'
            ,
            error: (xhr) =>
                @isSubmitting false
                app.error.Ajax(xhr)
        })

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = ->

    return {
        viewModel: VM
        template: templateMarkup
    }