$(document).ready(function() {
  "use strict";

  /**
   * Mark missing fields with an error.
   *
   * Returns true if all the required fields are non-empty.
   */
  function validate($form) {
    var missing = $form.find('input[required="required"]').filter(function() { return this.value === ""; });
    $.each(missing, function() {
      $(this).parent().addClass('error');
    });

    return missing.length === 0;
  }

  function showError(e) {
    $(document).find('div.attached.error.message').text(e.message).removeClass('hidden');
  }

  function formatPayload(json) {
    if (typeof json.env !== "undefined") {
      json.env = json.env.reduce(function(collector, element) {
        collector[element.key] = element.value;
        return collector;
      }, {});
    }

    json.task_cpus = parseFloat(json.task_cpus);
    json.task_mem = parseFloat(json.task_mem);

    return json;
  }

  function submitHandler(e) {
    var   $form = $('#new_task')
        , json = {}
        , env = []
        ;

    $form.find('div.error').removeClass('error');
    $form.find('div.attached.error.message').text('').addClass('hidden');
    e.preventDefault();

    $form.find('input[required="required"]')

    if (!validate($form)) {
      return null;
    };

    json = $form.serializeObject();
    json = formatPayload(json);

    $.ajax({
      method: 'POST',
      url: '/task',
      data: JSON.stringify(json),
      contentType: 'application/json',
      dataType: 'json',
      success: function(taskId) {
        window.location.href = window.location.origin + window.location.pathname + 'task/' + taskId;
      },
      error: function(xhr, e) {
        showError(xhr.responseJSON)
      }
    });
  }

  function createInput(type, number) {
    var input = {};
    if (type === 'env') {
      input = {
        one: {
          name: 'env['+number+'][key]',
          placeholder: 'key'
        },
        two: {
          name: 'env['+number+'][value]',
          placeholder: 'value'
        }
      }
    } else if (type === 'volumes') {
      input = {
        one: {
          name: 'volumes['+number+'][host_path]',
          placeholder: 'Host Volume'
        },
        two: {
          name: 'volumes['+number+'][container_path]',
          placeholder: 'Container Volume'
        }
      }
    } else {
      return $('<div />');
    }

    return $(
      '<div class="field ui action input ' + type + '">' +
        '<div class="two fields">' +
          '<div class="field">' +
            '<input name="' + input.one.name + '" placeholder="' + input.one.placeholder + '"/>' +
          '</div>' +
          '<div class="field">' +
            '<input name="' + input.two.name + '" placeholder="' + input.two.placeholder + '"/>' +
          '</div>' +
          '<button class="ui icon button">' +
            '<i class="minus red icon"></i>' +
          '</button>' +
        '</div>' +
      '</div>'
    );
  }

  function addVolumes(e) {
    var   $cont = $('#volumes')
        , index = $cont.data('count') + 1
        , $input = createInput('volumes', index)
        ;

    e.preventDefault();

    $cont.append($input);
    $cont.data('count', index);

  }

  function addEnvironments(e) {
    var   $cont = $('#env')
        , index = $cont.data('count') + 1
        , $input = createInput('env', index)
        ;

    e.preventDefault();

    $cont.append($input);
    $cont.data('count', index);
  }

  function addURIs(e) {
    var   $cont = $('#uris')
        , index = $cont.data('count') + 1
        , $input
        ;

    e.preventDefault();

    $input = $(
      '<div class="field ui action input uri">' +
        '<div class="field">' +
          '<input name="uri_' + index + '" placeholder="URI"/>' +
        '</div>' +
        '<button class="ui icon button">' +
          '<i class="minus red icon"></i>' +
        '</button>' +
      '</div>'
    );

    $cont.append($input);
    $cont.data('count', index);

  }

  function removeInput(e) {
    e.preventDefault();

    $(this).closest('div.ui.action.input').remove();
  }

  $('#new_task').on('submit', submitHandler);
  $('#new_task #submit').on('click', submitHandler);
  $('#new_task #volumes .plus').on('click', addVolumes);
  $('#new_task #env .plus').on('click', addEnvironments);
  $('#new_task #uris .plus').on('click', addURIs);
  $('#new_task #cancel').on('click', function(e) {
    e.preventDefault();
    window.location = window.location;
  });
  $('#new_task').on('keydown', 'input', function(e) {
    if (e.keyCode === 13) {
      e.preventDefault();
      if (validate($('#new_task'))) {
        $('#new_task').submit();
      }
    };
  });

  $(document).on('click', '.action.input button', removeInput);
})
