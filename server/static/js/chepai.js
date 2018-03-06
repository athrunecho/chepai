var timer;

function navigatePage(pageId) {
        $.mobile.navigate(pageId, {
        info: "info about the #bar hash"
    });
}

function unixTimeToStr(unixTime) {
    var unixTimestamp = new Date(unixTime * 1000); 
    return unixTimestamp.toLocaleString();
}

function getPublicInfo() {
    // Get time info
    $.ajax({
        type: "GET",
        url: "/api/unix_time_info",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/time_info" + " failed");
	    alert("获取时间信息失败");
        },
        success: function (data) {
            if (data.success) {
                $('#begin_time').text(unixTimeToStr(data.begin_time));
                $('#phase_one_end_time').text(unixTimeToStr(data.phase_one_end_time));
                $('#phase_two_end_time').text(unixTimeToStr(data.phase_two_end_time));
            } else {
                alert("获取时间信息失败: " + data.err);
            }
        },
        dataType: "json"
    });

    // Get license plate num
    $.ajax({
        type: "GET",
        url: "/api/license_plate_num",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/license_plate_num" + " failed");
	    alert("获取车牌数量失败");
        },
        success: function (data) {
            if (data.success) {
                $('#license_plate_num').text(data.license_plate_num);
            } else {
                alert("获取车牌数量失败: " + data.err);
            }
        },
        dataType: "json"
    });

    // Get start price
    $.ajax({
        type: "GET",
        url: "/api/start_price",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/start_price" + " failed");
	    alert("获取警示价失败");
        },
        success: function (data) {
            if (data.success) {
                $('#start_price').text(data.start_price);
            } else {
                alert("获取警示价失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function getBidderNum() {
    // Get bidder num
    $.ajax({
        type: "GET",
        url: "/api/bidder_num",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/bidder_num" + " failed");
	    clearInterval(timer);
	    alert("获取参拍人数失败");
        },
        success: function (data) {
            if (data.success) {
                $('#bidder_num').text(data.bidder_num);
            } else {
                alert("获取参拍人数失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function getLowestPrice() {
    // Get lowest price
    $.ajax({
        type: "GET",
        url: "/api/lowest_price",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/lowest_price" + " failed");
	    clearInterval(timer);
	    alert("获取当前最低成交价失败");
        },
        success: function (data) {
            if (data.success) {
                $('#lowest_price').text(data.lowest_price);
            } else {
                alert("获取当前最低成交价失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function getBidRecords() {
    // Get bid records
    $.ajax({
        type: "GET",
        url: "/api/bid_records",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/bid_records" + " failed");
            alert("获取出价记录失败");
        },
        success: function (data) {
            if (data.success) {
                // Clear records before update
                $('#bid_records').empty();
        
                $.each(data.bid_records, function(index, record) {
                     $('#bid_records').append('<div class="ui-body ui-body-a"><p>' + record + '</p></div>');
                });
            } else {
                alert("获取出价记录失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function getResult() {
    // Get result
    $.ajax({
        type: "GET",
        url: "/api/result",
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/result" + " failed");
            alert("获取成交结果失败");
        },
        success: function (data) {
            var msg;

            if (data.success) {
                if (data.done) {
                    msg = "恭喜你成交, 价格: " + data.price;
                } else {
                    msg = "你没有成交";
                }

                $('#result').text(msg);
            } else {
                alert("获取成交价格失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function login(id, password) {
    postData = {id: id, password: password};
    console.log(postData);

    $.ajax({
        type: "POST",
        url: "/api/login",
        data: JSON.stringify(postData),
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/login" + " failed");
        },
        success: function (data) {
            if (data.success) {
                //alert("登录成功\n" + "ID: " + id);
                getPublicInfo();
                navigatePage("#page2");
            } else {
                alert("提交失败: " + data.err);
            }
        },
        dataType: "json"
    });
}

function bid(price) {
    postData = {price: price};
    console.log(postData);

    $.ajax({
        type: "POST",
        url: "/api/bid",
        data: JSON.stringify(postData),
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            console.log("/api/price" + " failed");
            alert("出价失败");
        },
        success: function (data) {
	    var record;

            if (data.success) {
		    record = "第" + data.phase + "阶段出价成功: " + price;
            } else {
		    record = "第" + data.phase + "阶段出价失败: " + data.err;
            }

	    console.log(record);
	    $('#bid_records').append('<div class="ui-body ui-body-a"><p>' + record + '</p></div>');
	    
        },
        dataType: "json"
    });
}

$(document).ready(function () {
    //alert("document ready.");

    // Page 1 events.
    $('#loginBtn').click(function () {
        var id = $('#ID').val();
        var password = $('#password').val();

        login(id, password);
    });

    // Page 2 events.
    $('#bidBtn').click(function () {
        var price = Number($('#price').val());

        bid(price);
    });	

    $('#resultBtn').click(function () {
        getResult();
    });
});

$(document).on("pageinit","#page1",function(){
});

$(document).on("pagebeforeshow","#page1",function(){
});

$(document).on("pageinit","#page2",function(){
});

$(document).on("pagebeforeshow","#page2",function(){
    getPublicInfo();
    timer = window.setInterval(function() {
            getBidderNum();
            getLowestPrice();
    }, 1000);
});
