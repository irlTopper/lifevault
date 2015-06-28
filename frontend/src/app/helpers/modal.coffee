define ['knockout'], (ko) ->

    Modal = (app) ->
        @app = app
        return

    Modal::Init = (modalId,modalVM,params={},readyToShow=true) ->
        modalVM.templateLoaded = @app.ko.observable(false)
        modalVM.modalDivId = modalId + params.modalId
        modalVM.modalRef = params.modalRef
        if not modalVM.readyToShow?
            modalVM.readyToShow = ko.observable(readyToShow)
        modalVM.allowFade = ko.observable(true)#used to turn off fade-in/out effect class


        ShowModal = (modalVM) =>
            modalVM.modalDiv.modal("show")
            @app.currentModalVM = modalVM

            # Setup the after-closing-animation function:
            AttachOnhide(modalVM)

        AttachOnhide = (modalVM) =>
            modalVM.modalDiv.on('hidden.bs.modal', (e) =>

                # Clear the binding - not having this causes madness - Topper
                $( modalVM.modalDiv ).unbind( 'hidden.bs.modal' )

                # Invoke close callback if set
                if modalVM.modalRef.params.onCloseCallback?
                    modalVM.modalRef.params.onCloseCallback()

                # If not stacked, just close
                if not modalVM.stackedOnModalVM?
                    app.currentModalVM = undefined
                else# ..Otherwise restore the modal this was stacked on
                    ShowModal(modalVM.stackedOnModalVM)

                # Remove the modal component from the DOM by removing it in @app.modals
                @app.modals.remove(modalVM.modalRef)
            )

        onModalTemplateLoaded = () =>
            modalVM.modalDiv = $('#'+modalVM.modalDivId)
            modalVM.modalDiv.on('shown.bs.modal', (e) ->
                # Clear the binding - not having this causes madness - Topper
                $( modalVM.modalDiv ).unbind( 'shown.bs.modal' )
                # If there is an OnShow in the new modal, invoke it
                if modalVM.OnShow then modalVM.OnShow()
                # Make it draggable
                modalVM.modalDiv.find('.modal-dialog').draggable({
                    cancel: ".modal-body"
                })
            )
            if not@app.currentModalVM? # If no existing stack, just show the view model
                ShowModal(modalVM)
            else# Existing stack, need to hide existing modal and show ours

                # Here we drop the existing OnHide (hidden.bs.modal) handler,
                # and replace it with one that will have no animation..
                # Then once we've hidden the modal and shown the new one
                # we re-attach the original OnHide function - which is important

                # Clear the existing binding - not having this causes madness - Topper
                $( app.currentModalVM.modalDiv ).unbind( 'hidden.bs.modal' )

                # Create the new onHide ('hidden.bs.modal') handler
                app.currentModalVM.modalDiv.on('hidden.bs.modal', (e) ->
                    # Clear the binding - not having this causes madness - Topper
                    $( app.currentModalVM.modalDiv ).unbind( 'hidden.bs.modal' )
                    # Tell the new modal that it is stacked on another so we can restore
                    modalVM.stackedOnModalVM = app.currentModalVM
                    # Turn off the fade-in effect
                    modalVM.allowFade(false)
                    # Show the new modal
                    ShowModal(modalVM)
                )
                # Turn off the fade-out effect
                app.currentModalVM.allowFade(false)
                # Hide the current modal
                app.currentModalVM.modalDiv.modal("hide")


        subscription = modalVM.templateLoaded.subscribe (isTemplateLoaded) ->
            if isTemplateLoaded
                onModalTemplateLoaded()
                # Remove this subscription
                subscription.dispose()
                subscription = undefined
        return



    # Wrapper function for showing a modal
    Modal::Confirm = (title, question, callback) ->
        @app.modal.Show("confirm",{
            title: title
            question: question
            callback: callback
        },@)
        return

    # Loads a modal by adding it to the @app.modals observable array
    # Also ensures a modal is not loaded twice
    Modal::Show = (name,params={},parentVM) ->

        # Create something to hold this reference
        ModalRef = (name,params) ->
            @name = name
            params.parentVM = parentVM
            @params = params
            @readyToShow = false
            return @

        modalRef = new ModalRef(name,params)

        # Setup the loaded callback - modals are expected to call this
        $.extend(params, {
            modalId: @app.modals().length+1
            modalRef: modalRef
        })

        @app.modals.push modalRef
        return

    Modal::SetOnCloseCallback = (modalVM, fn) ->
        modalVM.modalRef.params.onCloseCallback = fn


    Modal::Close = (closingModalVM, runCallback = true) ->
        if closingModalVM.modalRef.params.callback? and runCallback
            closingModalVM.modalRef.params.callback()

        # Close it
        closingModalVM.modalDiv.modal("hide")
        return

    return Modal
