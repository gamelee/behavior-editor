//把表单转换出json对象
$.fn.toJson = function () {
    let self = this,
        json = {},
        push_counters = {},
        patterns = {
            "validate": /^[a-zA-Z][a-zA-Z0-9_]*(?:\[(?:\d*|[a-zA-Z0-9_]+)\])*$/,
            "key": /[a-zA-Z0-9_]+|(?=\[\])/g,
            "push": /^$/,
            "fixed": /^\d+$/,
            "named": /^[a-zA-Z0-9_]+$/
        };

    this.build = function (base, key, value) {
        base[key] = value;
        return base;
    };

    this.push_counter = function (key) {
        if (push_counters[key] === undefined) {
            push_counters[key] = 0;
        }
        return push_counters[key]++;
    };

    $.each($(this).serializeArray(), function () {
        // skip invalid keys
        if (!patterns.validate.test(this.name)) {
            return;
        }

        let k,
            keys = this.name.match(patterns.key),
            merge = this.value,
            reverse_key = this.name;

        while ((k = keys.pop()) !== undefined) {
            // adjust reverse_key
            reverse_key = reverse_key.replace(new RegExp("\\[" + k + "\\]$"), '');
            // push
            if (k.match(patterns.push)) {
                merge = self.build([], self.push_counter(reverse_key), merge);
            }
            // fixed
            else if (k.match(patterns.fixed)) {
                merge = self.build([], k, merge);
            }
            // named
            else if (k.match(patterns.named)) {
                merge = self.build({}, k, merge);
            }
        }
        json = $.extend(true, json, merge);
    });

    return json;
};

//将josn对象赋值给form
$.fn.loadData = function (obj) {
    let key, value, tagName, type, arr;
    this.get(0).reset();

    for (let x in obj) {
        if (obj.hasOwnProperty(x)) {
            key = x;
            value = obj[x];

            this.find("[name='" + key + "'],[name='" + key + "[]']").each(function () {
                tagName = $(this)[0].tagName.toUpperCase();
                type = $(this).attr('type');
                if (tagName === 'INPUT') {
                    if (type === 'radio') {
                        if ($(this).val() === value) {
                            $(this).attr('checked', true);
                        }
                    } else if (type === 'checkbox') {
                        arr = value.split(',');
                        for (let i = 0; i < arr.length; i++) {
                            if ($(this).val() === arr[i]) {
                                $(this).attr('checked', true);
                                break;
                            }
                        }
                    } else {
                        $(this).val(value);
                    }
                } else if (tagName === 'SELECT' || tagName === 'TEXTAREA') {
                    $(this).val(value);
                }
            });
        }
    }
}