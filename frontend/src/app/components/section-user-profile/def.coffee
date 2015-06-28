define [
    "knockout"
    "text!./template.html"
    "jquery.validate"
    "userModel"
    "lodash"
], (ko, templateMarkup, jqueryvalidate, userModel, _) ->
    VM = (params) ->
        @user = params.user
        @readyToShow = ko.observable false
        @isSubmitting = ko.observable false
        @showAlternateEmails = ko.observable false
        @timeGroups = ko.observable {}
        @indicators = ko.observable {}

        $.when(
            $.getJSON "v1/users/" + @user.id() + ".json", (response,d,xhr) =>
                #editUser is the copy we will use for editing
                @editUser = app.mapping.fromJS(response).user
                @userModel = ko.observable(app.FindUserById(@editUser.id()))
                @editUser.changePassword = ko.observable false
                @editUser.password = ko.observable ""
                @editUser.confirmPassword = ko.observable ""

                altEmailsOrigLen = @editUser.altEmails().length
                while @editUser.altEmails().length < 3 then @editUser.altEmails.push("")
                if altEmailsOrigLen > 0 then @ShowAlternateEmails()

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)
        ,

            $.getJSON "v1/timezones.json", (response,d,xhr) =>
                @timezones = app.mapping.fromJS(response).timezones
                @timeGroups(_.uniq(@timezones(), (zone) =>
                    zone.offsetDisplay()
                ))

                for group in @timeGroups()
                    group.zones = _.filter(@timezones(), (zone) =>
                        zone.offsetDisplay() == group.offsetDisplay()
                    )

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)
        ,

            $.getJSON "v1/timeformats.json", (response,d,xhr) =>
                @timeformats = app.mapping.fromJS(response).timeformats

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)
        ,

            $.getJSON "v1/users/phone/indicator.json", (response,d,xhr) =>
                @indicators = app.mapping.fromJS(response)
                @defaultIndicator = {
                    countryCode: ko.observable ""
                    label: ko.observable "Select your country..."
                }
                @indicators.unshift(@defaultIndicator)

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)

        ).then =>
            @readyToShow true



        app.InitOnShow(@)# Sets-up the OnShow() function
        return


    VM::OnShow = () ->
        # Setup validation
        $('#userProfileForm').validate({onsubmit: false})
        # Focus first field
        $("#userProfileForm input[type=text]").first().focus()
        # Setup 'pops' if any
        $('.pops').popover({ trigger:"hover", html:true })
        $('.pops-link').popover({ trigger:"hover", html:true, delay: { "hide": 1000 }})
        return

    VM::ChangePhoto = () ->
        app.modal.Show 'imageupload', {
            id: @editUser.id(),
            fileType:'userimage',
            callback: (newPhotoURL) =>
                @onChangePhoto(newPhotoURL)
            }, @
        return

    VM::onChangePhoto = (newPhotoURL) ->
        @editUser.avatarURL(newPhotoURL)

        if @editUser.id() is app.loggedInUser.id()
            app.loggedInUser.avatarURL(newPhotoURL)
            app.SaveLoggedInUser()

        @user.avatarURL(newPhotoURL)

        # Send touch request to TWPM
        $.ajax({
            url: "../people/" + @editUser.id() + "/touch.json",
            type: "PUT",
            dataType: 'json'
        })

        return

    VM::DeleteUserImage = (o) ->
        @onChangePhoto("")
        @OnClickSaveUserProfile()
        return

    VM::HasAnAvatar = () ->
        a = @editUser.avatarURL()
        matches = new RegExp("images\/avatars\/((fe)?male|unknown)\/avatar")
        return a != "" and not matches.test(a)


    VM::ResetPassword = (o) ->
        q = "Are you sure you want to reset " + @editUser.firstName() + "'s password?"
        app.modal.Confirm "Please confirm", q, =>
            @isSubmitting true
            url = 'v1/users/' + @editUser.id() + '/resetpassword.json'
            $.ajax({
                url: url, type: 'POST',
                dataType: 'json',
                success: (response,d,xhr) =>
                    app.flash.Success("<strong>Done!</strong> " + @editUser.firstName() + " had a password reset sent to them.", {timer: 800})
                    @isSubmitting false
                ,
                error: (xhr) =>
                    @isSubmitting false
                    app.error.Ajax(xhr)
            })

        return


    VM::DeleteUser = (o) ->
        q = "Are you sure you want to delete " + @editUser.firstName() + "?"
        app.modal.Confirm "Please confirm", q, =>
            @isSubmitting true
            url = 'v1/users/' + @editUser.id() + '.json'
            $.ajax({
                url: url, type: 'DELETE',
                dataType: 'json',
                success: (response,d,xhr) =>
                    app.flash.Success("<strong>Done!</strong> " + @editUser.firstName() + " was deleted.")
                    app.UpdateUsers()
                    app.GoTo "settings/users"
                ,
                error: (xhr) =>
                    @isSubmitting false
                    app.error.Ajax(xhr)
            })

        return


    VM::ShowAlternateEmails = (o) ->
        @showAlternateEmails(true)
        $('.alternate-emails').first().focus()
        return

    VM::IsSelected = (i) ->
        false


    VM::OnClickSaveUserProfile = () ->
        @isSubmitting true

        if not $('#userProfileForm').valid()
            app.flash.Error("<strong>Oh no</strong> &mdash; You have some missing or incorrect values.", {timer:800} )
            @isSubmitting false
            return

        if @editUser.changePassword() and @editUser.confirmPassword() isnt @editUser.password()
            app.flash.Error("<strong>Oh no</strong> &mdash; Your passwords don't match")
            @isSubmitting false
            return false

        $.ajax({
            url: "v1/users/" + @user.id() + ".json",
            type: 'PUT',
            dataType: 'json',
            data: $('#userProfileForm').serialize()

            success: (response,d,xhr) =>
                @isSubmitting false
                if @editUser.id() is app.loggedInUser.id()
                    app.flash.Success("Your profile has been updated.", {timer:500 })
                else
                    app.flash.Success("The users's profile was saved successfully.", {timer:500 })
            ,
            error: (xhr) =>
                @isSubmitting false
                app.error.Ajax(xhr)
        })
        return

    return {
        viewModel: VM
        template: templateMarkup
    }