Teamwork Desk
===

## Development environment setup guide

### Prerequisites
- [Node](http://nodejs.org/)
- [EditorConfig](http://editorconfig.org) plugin for your IDE / text-editor
- `npm i -g gulp` to install [Gulp](http://gulpjs.com/) globally
- `npm i -g bower` to install [Bower](http://bower.io/) globally

### Dependencies
- `npm i` to install local node dependencies
- `bower i` to install local bower dependencies

### Commands
- `gulp` compiles the code.
- `gulp open` will compile the code & open it in your browser (and sync clicks, scrolls, etc). No need for a webserver of your own.
- `gulp watch` will do the same as `gulp open` but montiors your files for changes and injects the compiled output into the browser.
- `gulp docs` will generate [Biscotto](https://github.com/atom/biscotto) docs and open them in your browser.


## Config
Copy this to config.js in src/

- window.twDeskConfig = {
-     "isDeveloperMode": true,
-     "APIURL": "http://localhost:9000/"
- }


## Internal class names for ViewModels and Models

ViewModels for components should all use the class name "VM" internally.
This allows us to refactor code quickly and it doesn't matter because we never see this name.
eg. app/components/modal-confirm/def.coffee uses "VM" internally
ie. All components should use "VM" for their class name.

Models on the other hand should always have a proper name - because they can listed
in the console and it's useful for context and debugging.
eg. helpDocsSiteModel.coffee uses "HelpDocsSiteModel"



## Using "Help Docs" instead of "Articles"

I've decided to rename "Articles" to "Help Docs" because it just seems more right.

## One default "Help Site" that can't be deleted

There will always be at least one "Help Site" but it doesn't have to be published.



## API Calls

All the API calls are listed in [routes](https://github.com/irlTopper/ohlife2/blob/master/conf/routes)


## Updating a ticket

Can make changes and use just SaveChanges if you know the field that has changed (you do)
        @ticket.status( newStatus )
        @ticket.SaveChanges(["status"])


or track the changes to several fields automatically and have one ajax update with:
        @ticket.TrackChanges () =>
            @ticket.status( newStatus )
            @ticket.subject( "Topper is awesome" )
        .SaveChanges()


## Tips

### Beware of bootstrap event binding

If you want to perform an action after a boostrap animation like
hiding a modal with a subscription to 'hidden.bs.modal', it is essential
that you clear the bind afterwards... or you'll get
all sorts of madness. You can clear the bind in the handler with:
$( ELEMENT ).unbind( 'hidden.bs.modal' )

### Installing shite from bower
From the "frontend" folder root - run
bower install SOMETHING --save

The --save bit is important - it updates the bower.json file.


## Coffeewatch
Rob cam up with the command
gulp coffeewatch
It just watches for changes to the coffeescript and compiles the javascript
in light speed. It's the business.

# Modal confirm
Works just like a javascript confirm() but prettier
Type this in console:
app.ShowModal("prompt",{ title: "Merge", question: "What should be called the tag?", callback: function(a){alert(a)}   },window)


# Modal loading system
OK, we have a nice on-demand modal loading and stacking system in TeamworkDesk.
We can just call say
app.ShowModal "add-user", {}, this
and the app will:

1. Load the viewModel and template for the modal. A new hidden component is added to the page.
2. Once the template is loaded, the params you pass (eg. {} above) will be passed to the modal viewModel new instance..
3. The modal is automatically shown and the OnShow method of the modal will is called

How the magic works...

ShowModal creates a ModalRef object and pushes this to an observableArray.
The observableArray is plugged into the index.html template - it created an instance
of each modal component.
	Each modal has to has to call app.ModalInit('adduser', @, params) in it's constructor.
(The first argument doesn't really matter - just for nice div id's)

What this does is setups up a whole raft of logic that waits
until the component template is loaded (see templateLoaded system), then
displays the bootstrap component (using modalDiv.modal("show")),
watches for the shown event (shown.bs.modal), triggers the OnShow if there is one.

It also sets up something to watch "hide" modal events and automatically remove the
component from the DOM and memory by removing it from the "modals" observableArray
after the closing animation.

Finally it was extended to allow for stacking of modals; it does this by
detecting if there is a stack when adding a new modal.. and if there is, it:

1. Removes the existing "onclose" handler.
2. Closes the current modal with no animations (but not deleting)
3. Showing the new modal (with no animation).
4. Later when the new modal is closed, it checks to see if there is an existing stack and if so, shows the previous modal again and sets up the onclose event on it again.

It's complex yes, but it works very well now.


