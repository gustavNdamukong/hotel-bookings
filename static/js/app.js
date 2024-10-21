
function Prompt() {
    let toast = function (c) {
        const {
            msg = '',
            icon = 'success',
            position = 'top-end',

        } = c;

        // 'didOpen' is a Sweet alert (Swal) life-cycle hook that will run after the modal displays
        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
            }
        })

        Toast.fire({})
    }

    let success = function (c) {
        const {
            msg = "",
            title = "",
            footer = ""
        } = c

        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    let error = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c

        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    async function custom(c) {
        const {
            icon = "",
            msg = "",
            title = "",
            showConfirmButton = true,
        } = c;

        //'willOpen' is another Sweet alert (Swal) life-cycle hook/function that will run before the alert modal opens
        //'didOpen' is another (life-cycle hook) that will run after the modal displays
        const {value: result} = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            willOpen: () => {
                if (c.willOpen !== undefined) {
                    c.willOpen();
                }
            },
            didOpen: () => {
                if (c.didOpen !== undefined) {
                    c.didOpen();
                }
            }
        })

        if (result) {
            if (result.dismiss !== Swal.DismissReason.cancel) {
                if (result.value !== "") {
                    if (c.callback !== undefined) {
                        c.callback(result);
                    }
                } else {
                    c.callback(false);
                }
            } else {
                c.callback(false);
            }
        }
    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}

function BookARoom(roomID, token) {
    console.log("calling BookARoom() with roomID: "+roomID+" Token: "+token);
    let html = `
      <form id="check-availability-form" action="" method="post" novalidate class="needs-validation">
          <div class="form-row">
              <div class="col">
                  <div class="form-row" id="reservation-dates-modal">
                      <div class="col">
                          <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">
                      </div>
                      <div class="col">
                          <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                      </div>

                  </div>
              </div>
          </div>
      </form>
      `;

    //  showOnFocus: true, (means when you click on it, you should see it)
    //  autohide: true, (whether to hide date picker once a date is selected)
    //  minDate: new Date(), (do not allow dates in the past)
    attention.custom({
        title: 'Choose your dates',
        msg: html,

        title: 'Choose your dates',
        msg: html,
        willOpen: () => {
            const elem = document.getElementById("reservation-dates-modal");
            const rp = new DateRangePicker(elem, {
                format: 'yyyy-mm-dd',
                showOnFocus: true,
                minDate: new Date(),
            })
        },
        didOpen: () => {
            //remove the disabled class on input fields after alert popup opens 
            //we added the disabled class coz we did not want the datepicker to 
            //auto-open on the popup load, since sweet alert auto-sets the first 
            //form field on focus, causing the datepicker to open unprompted
            document.getElementById("start").removeAttribute("disabled");
            document.getElementById("end").removeAttribute("disabled");
        },


        callback: function(result) {
            let form = document.getElementById("check-availability-form");
            let formData = new FormData(form);
            formData.append("csrf_token", token);
            formData.append("room_id", roomID);

            fetch('/search-availability-json', {
                method: "post",
                body: formData,
            })
            .then(response => response.json())
            .then(data => {
                if (data.ok) {
                    attention.custom({
                        icon: "success",
                        showConfirmButton: false,
                        msg: '<p>Room is available!</p>'
                            + '<p><a href="/book-room?id='
                            + data.room_id
                            + '&s='
                            + data.start_date
                            + '&e='
                            + data.end_date
                            + '" class="btn btn-primary">'
                            + 'Book now!</a></p>',
                    })
                }
                else
                {
                    attention.error({
                    msg: "Not available",
                    })
                }
            })
        }
    });
}


/*/--------------------------------------------------------------------------------------------------------------------//
                            NOTES ON HOW TO USE THE notie & Sweet Alert JS LIBRARIES FOR NOTIFICATIONS
//--------------------------------------------------------------------------------------------------------------------/
-Note that the two alert libraries are pulled in from:

    <link rel="stylesheet" type="text/css" href="https://unpkg.com/notie/dist/notie.min.css">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/sweetalert2@10.15.5/dist/sweetalert2.min.css">

-Once these links are referenced, you will now have available the methods notie and Slal which you can call like so:

    notie.alert({...the required parameters here...})

    Swal.mixin({...optionally do configuration settings here before calling Swal.fire({})...})
    Swal.fire(({...the required parameters here...})
*/



/*
function alertSuccess(msg) {
notie.alert({
    type: success (optional-default = 4), // the options are: enum: [1, 2, 3, 4, 5, 'success', 'warning', 'error', 'info', 'neutral']
    text: msg,
    //stay: Boolean (optional-default = false), // (should it stay on screen or not)
    //time: Number (optional-default = 3, minimum = 1), // (how long the popup will stay on screen for-in secs)
    //position: String (optional-default = 'top'), // the options are: enum: ['top', 'bottom']
})
} */

//Examples of how to use the Sweet Alert (Swal) via the Prompt() function above
//---------------------------------------------------------------------------
/*
    // We can create custom funcs to wrap calls to notie.alert({}) and Swal.fire({}) so we can make them generic

    //See Prompt() function above
    let attention = Prompt();

    function notify(msg, msgType) {
        notie.alert({
            type: msgType,
            text: msg,
        })
    }

    function notifyModal(title, text, icon, confirmationButtonText) {
        Swal.fire({
            title: title,
            html: text,
            icon: icon,
            confirmButtonText: confirmationButtonText
        })
    }

    notify('This is my message', 'warning');
    notifyModal('Some title', '<em>Hello world</em>', 'success', 'my Text for the button');
    attention.toast({ msg: "Hello, world", icon: "error" });
    attention.success({ msg: "Hello, world", footer: "This is the footer"});
    attention.error({ msg: "Oops, something went wrong", footer: "This is the footer"});
*/