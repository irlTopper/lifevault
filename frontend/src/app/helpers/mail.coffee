define ["knockout"], (ko) ->

    Mail = ->

    Mail::OpenMailModal = (inbox, forwardEmailsTo) ->
        @forwardEmailsTo = forwardEmailsTo
        app.modal.Show("send-email",
            {
                id: "forward-email",
                inboxId: inbox.id(),
                title: "Email Forwarding Instructions",
                subject: "" + app.utility.SafeHTML(app.loggedInUser.firstName() + " " + app.loggedInUser.lastName()) + " wants you to forward #{inbox.email()} to LifeVault",
                message: @ForwardEmail(inbox)
            }
            , @)
        return

    Mail::ForwardEmail = (inbox) ->
        message = " Hi!<br /><br />
                    I'm in the process of setting up a LifeVault (see www.teamwork.com/desk) to help manage our customer support emails and I need your help.<br /><br />
                    Can you please create forward all emails sent to <strong>#{inbox.email()}</strong>  to <strong>#{@forwardEmailsTo}</strong>?<br /><br />
                    The support team at LifeVault have been CC'd here in case you have any follow-up questions.<br /><br />
                    Thank you! <br /><br />
                    #{app.loggedInUser.firstName()} #{app.loggedInUser.lastName()}
                    "
        return message

    Mail::OpenHelpMailModal = ->
        app.modal.Show("add-support-ticket",
            {
                to: "desk@teamwork.com",
                subject: "Please help me set up the inbox forwarding email"
                title: "Need help ?"
                message: "Hi LifeVault support, I would appreciate some help setting up the inbox forwarding."
            }
            , @)
        return

    return Mail