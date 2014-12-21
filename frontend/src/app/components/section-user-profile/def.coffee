define [
    "knockout"
    "text!./template.html"
    "jquery.validate"
], (ko, templateMarkup, jqueryvalidate) ->
    VM = (params) ->
        @user = params.user
        @zzz = params.user
        @isLoaded = ko.observable false
        @isSubmitting = ko.observable false
        @showAlternateEmails = ko.observable false

        $.when(
            $.getJSON "v1/users/" + @user.id() + ".json", (response,d,xhr) =>
                #editUser is the copy we will use for editing
                @editUser = app.mapping.fromJS(response).user
                @editUser.changePassword = ko.observable false
                @editUser.password = ko.observable ""
                @editUser.confirmPassword = ko.observable ""
                if @editUser.altEmails().length > 0 then @ShowAlternateEmails()

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)
        ,

            $.getJSON "v1/timezones.json", (response,d,xhr) =>
                @timezones = app.mapping.fromJS(response).timezones

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)
        ,

            $.getJSON "v1/timeformats.json", (response,d,xhr) =>
                @timeformats = app.mapping.fromJS(response).timeformats

            .fail (xhr) ->
                return app.showErrorLoadingMsg(xhr)

        ).then =>
            @isLoaded true



        app.InitOnShow(@)# Sets-up the OnShow() function
        return


    VM::OnShow = () ->
        # Setup validation
        $('#userProfileForm').validate({onsubmit: false})
        # Focus first field
        $("#userProfileForm input[type=text]").first().focus()
        return


    VM::ChangePhoto = () ->
        app.ShowModal 'imageupload', {id: @editUser.id(), fileType:'userimage', callback: (newPhotoURL) =>
            @onChangePhoto(newPhotoURL)
        }, @

    VM::onChangePhoto = (newPhotoURL) ->
        @editUser.avatarURL(newPhotoURL)

        if @editUser.id() is app.loggedInUser.id()
            app.loggedInUser.avatarURL(newPhotoURL)
            app.SaveLoggedInUser()

        @user.avatarURL(newPhotoURL)



    VM::DeleteUserImage = (o) ->
        @editUser.avatarURL("")
        return

    VM::HasAnAvatar = () ->
        a = @editUser.avatarURL()
        return a != "" and a.indexOf("/images/avatars/avatar") == -1


    VM::ResetPassword = (o) ->
        q = "Are you sure you want to reset " + @editUser.firstName() + "'s password?"
        app.Confirm "Please confirm", q, =>
            @isSubmitting true
            url = 'v1/users/' + @editUser.id() + '/resetpassword.json'
            $.ajax({
                url: url, type: 'POST',
                dataType: 'json',
                success: (response,d,xhr) =>
                    app.FlashSuccess("<strong>Done!</strong> " + @editUser.firstName() + " had a password reset sent to them.", {timer: 800})
                    @isSubmitting false
                ,
                error: (xhr) =>
                    @isSubmitting false
                    app.HandleAjaxError(xhr)
            })

        return


    VM::DeleteUser = (o) ->
        q = "Are you sure you want to delete " + @editUser.firstName() + "?"
        app.Confirm "Please confirm", q, =>
            @isSubmitting true
            url = 'v1/users/' + @editUser.id() + '.json'
            $.ajax({
                url: url, type: 'DELETE',
                dataType: 'json',
                success: (response,d,xhr) =>
                    app.FlashSuccess("<strong>Done!</strong> " + @editUser.firstName() + " was deleted.")
                    app.UpdateUsers()
                    app.GoTo "settings/users"
                ,
                error: (xhr) =>
                    @isSubmitting false
                    app.HandleAjaxError(xhr)
            })

        return


    VM::ShowAlternateEmails = (o) ->
        @showAlternateEmails(true)
        while @editUser.altEmails().length < 3 then @editUser.altEmails.push("")
        return


    VM::OnSaveUserProfile = (o) ->
        @isSubmitting true

        if not $('#userProfileForm').valid()
            app.FlashError("<strong>Oh no</strong> &mdash; You have some missing or incorrect values.", {timer:800} )
            return

        if @editUser.changePassword() and @editUser.confirmPassword() isnt @editUser.password()
            app.FlashError("<strong>Oh no</strong> &mdash; Your passwords don't match")
            @isSubmitting false
            return false

        $.ajax({
            url: "v1/users/" + @user.id() + ".json",
            type: 'PUT',
            dataType: 'json',
            data: $('#userProfileForm').serialize()
            success: (response,d,xhr) =>
                @isSubmitting false
                app.FlashSuccess("The users's settings were saved!", {timer:500 })
                app.UpdateUsers()
            ,
            error: (xhr) =>
                @isSubmitting false
                app.HandleAjaxError(xhr)
        })
        return

    return {
        viewModel: VM
        template: templateMarkup
    }