$(document).ready(function() {
  "use strict";

  var taskId = $('body').data('task');

  function format(data) {
    data = data.split("\n");

    var $el = $('#stdout'),
        cls = 'gray',
        showNext = false;
    $.each(data, function(i, v) {
      if (showNext && cls == 'gray') {
        cls = '';
      }
      if (showNext) {
        $el.append($('<p/>').append(ansi_up.ansi_to_html(v)));
      } else {
        $el.append($('<p/>', { text: v, class: cls }));
      }

      if (v.indexOf("Starting task") == 0) {
        showNext = true;
      }
    })
    $el.find('.gray').hide();
    $('#show_stdout').text($el.find('.gray').length + ' lines hidden. Click to show.');
  }

  function getLogs(logfile) {
    var $el = $("#" + logfile);
    if (!$el) {
      return;
    }
    $.ajax({
      method: 'GET',
      url: '/task/' + taskId + '/' + logfile,
      success: function(data) {
        if (typeof data === 'undefined') {
          return
        }
        if (data.length == 0) {
          $('div.logs').hide();
          return;
        }
        $el.text('');
        if (logfile === 'stdout') {
          format(data);
        } else {
          $.each(data.split("\n"), function(i, v) {
            $el.append($('<p/>', { text: v, class: 'gray' }));
          });
        };
      },
      error: function(xhr, e) {
        $el.text(e)
      }
    });
  }

  $('body').on('click', '#kill', function(e) {
    e.preventDefault();
    $.ajax({
      method: 'POST',
      url: '/task/' + taskId + '/kill',
      success: function() {
        window.location = window.location;
      },
      error: function(xhr, e) {
        $('.error.hidden').removeClass('hidden')
        $('.error .information').text(xhr.responseText)
      }
    });
  })

  $('body').on('click', '.close', function(e) {
    e.preventDefault();
    $(this).parents('.ui.error').addClass('hidden')
  })

  $('body').on('click', '#show_stdout', function(e) {
    e.preventDefault();
    $('#stdout p.gray').show();
    $('#show_stdout').remove();
  })

  getLogs('stdout');
  getLogs('stderr');
})
