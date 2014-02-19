// TagBBS module
var TagBBS = angular.module("TagBBS", ["ngRoute", "ngProgressLite"]);

TagBBS.config(function($routeProvider, $locationProvider) {
    $routeProvider
    .when("/_/login", {
        templateUrl: "login.html",
        controller: "Login"
    })
    .when("/_/logout", {
        templateUrl: "logout.html",
        controller: "Logout"
    })
    .when("/_/register", {
        templateUrl: "register.html",
        controller: "Register"
    })
    .when("/_/passwd", {
        templateUrl: "passwd.html",
        controller: "Register"
    })
    .when("/:/:key?", {
        templateUrl: "put.html",
        controller: "Put"
    })
    .when("/:query/:key", {
        templateUrl: "show.html",
        controller: "Show"
    })
    .when("/:query", {
        templateUrl: "list.html",
        controller: "List",
        reloadOnSearch: false
    })
    .otherwise({
        templateUrl: "404.html"
    });

    $locationProvider.html5Mode(false);
})
.controller("MainCtrl", function($scope, $location, $route, bbs) {
    $scope.user = "";
    $scope.homepage = "/@/post:0";
    $scope.setUser = function(user) {
        $scope.user = user;
    };
    if (!bbs.session()) {
        var metapath = $location.path().split("/")[1] == "_";
        if (!metapath) {
            localStorage.returnPath = $location.url();
            $location.url("/_/login");
        } else {
            localStorage.returnPath = "";
        }
    }
    bbs.version().success(function(d) {
        $scope.name = d.result.name;
        $scope.version = d.result.version;
    });
})
.controller("Login", function($scope, $location, bbs) {
    $scope.user = "";
    $scope.pass = "";
    $scope.pass2 = "";
    $scope.message = "";
    var redirect = function(d) {
        if (d.result) {
            if (localStorage.returnPath) {
                $location.url(localStorage.returnPath);
                localStorage.returnPath = "";
            } else {
                $location.url($scope.homepage);
            }
        }
    };
    $scope.submit = function() {
        bbs.login($scope.user, $scope.pass).success(function(d) {
            if (d.result) {
                localStorage.sid = d.result;
                $scope.setUser($scope.user);
            } else {
                $scope.message = "Login failed: " + d.error;
            }
        }).success(redirect);
    };

    if (localStorage.sid) {
        $scope.message = "Existing session detected, checking...";
        bbs.session(localStorage.sid);
        bbs.who().success(function(d) {
            if (d.error) {
                $scope.message = "Existing session not valid: " + d.error;
            }
        }).success(function(d) {
            $scope.setUser(d.result);
            redirect(d);
        });
    }
})
.controller("Logout", function($scope, $location, bbs) {
    $scope.error = "logging out...";
    bbs.logout().success(function(d) {
        if (d.error) {
            $scope.error = d.error;
        } else {
            $location.url("/_/login");
        }
        localStorage.sid = "";
        $scope.setUser("");
    });
})
.controller("Register", function($scope, $location, bbs) {
    $scope.user = "";
    $scope.oldpass = "";
    $scope.pass = "";
    $scope.pass2 = "";

    $scope.message = "";
    $scope.submit = function() {
        if ($scope.pass != $scope.pass2) {
            $scope.message = "Password Mismatch!";
            return;
        }
        bbs.login($scope.user, $scope.oldpass).success(function(d) {
            if (d.result) {
                localStorage.sid = d.result;
                $scope.setUser($scope.user);
                bbs.register().success(function() {
                    bbs.passwd($scope.pass).success(function(d) {
                        if (d.error) {
                            $scope.message = "Error setting password: " + d.error;
                        }
                        $location.url($scope.homepage);
                    })
                })
            } else {
                if ($scope.oldpass) {
                    $scope.message = "Wrong User or old Password!";
                } else {
                    $scope.message = "Registration failed, probably this id has been taken.";
                }
                console.log(d.error);
            }
        })
    };
})
.controller("List", function($scope, $routeParams, $location, $route, $timeout, bbs) {
    $scope.query = $routeParams.query || "";
    $scope.tags = [];
    $scope.posts = [];

    $scope.newHandQuery = function() {
        $location.url("/" + $scope.query);
        $route.reload();
    };
    $scope.put = function() {
        $location.url("/:").search({tags: $scope.tags});
    };

    var parsed = {}; // dirty hack to prevent double refresh
    var jump = function(cursor, before, after) {
        var parts = [];
        if ($scope.query) parts.push($scope.query);
        if (cursor) parts.push("@" + cursor);
        if (before) parts.push("-" + before);
        if (after) parts.push("+" + after);
        bbs.list(parts.join(" ")).success(function(d) {
            if (!d.result) return;
            parsed = d.result.query;
            $scope.query = parsed.tags.join(" ");
            $scope.tags = parsed.tags;
            if (d.result.posts.length == 0) {
                if ($scope.posts.length > 0) {
                    $scope.message = "No more...";
                } else {
                    $scope.message = "No result...";
                }
                $timeout(function() {
                    $scope.message = "";
                }, 2000);
                return;
            } else {
                $scope.message = null;
            }
            $scope.posts = d.result.posts;
            // update search if not default
            if (parsed.cursor || parsed.before != 20 || parsed.after != 0) {
                $location.search({
                    cursor: parsed.cursor,
                    before: parsed.before,
                    after: parsed.after
                });
            }
        });
    };
    $scope.firstPage = function() {
        jump("", 0, 20);
    };
    $scope.lastPage = function() {
        jump("", 20, 0);
    };
    $scope.prevPage = function() {
        if ($scope.posts.length > 0) {
            jump($scope.posts[0].key, 20, 0);
        }
    };
    $scope.nextPage = function() {
        if ($scope.posts.length > 0) {
            jump($scope.posts[$scope.posts.length-1].key, -1, 21);
        }
    };
    $scope.$on('$routeUpdate', function() {
        if (parsed.cursor == $routeParams.cursor && parsed.before == $routeParams.before && parsed.after == $routeParams.after) return;
        jump($routeParams.cursor, $routeParams.before, $routeParams.after);
    });
    jump($routeParams.cursor, $routeParams.before, $routeParams.after);
})
.controller("Show", function($scope, $routeParams, $location, $timeout, bbs) {
    $scope.key = $routeParams.key;
    $scope.message = "";
    $scope.post = {};
    $scope.loading = true;

    $scope.show_raw = function() {
        window.open('data:text/plain;charset=utf-8,' + encodeURIComponent($scope.post.content));
    };
    $scope.reply = function() {
        var post = $scope.post;
        $location.url("/:").search({
            title: post.header.title.substring(0,3) == 'Re:' && post.header.title || 'Re: ' + post.header.title,
            reply: $scope.key,
            thread: post.header.thread || $scope.key,
            tags: post.header.tags || [],
        });
    };
    $scope.query = function() {
        if ($routeParams.query == '@') return null;
        return $routeParams.query;
    };
    $scope.thread = function() {
        return $scope.post.header && $scope.post.header.thread || $scope.key;
    };
    // TODO cache the list
    $scope.prev = function() {
        var q = $scope.query();
        bbs.list(q + " @" + $scope.key + " -10").success(function(d) {
            var posts = d.result && d.result.posts || [];
            if (posts.length == 0) {
                $scope.message = "the end of the list...";
                $timeout(function() {
                    $scope.message = "";
                }, 2000);
                return;
            } else {
                $location.url("/" + q + "/" + posts[posts.length-1].key);
            }
        });
    };
    $scope.next = function() {
        var q = $scope.query();
        bbs.list(q + " @" + $scope.key + " --1 +11").success(function(d) {
            var posts = d.result && d.result.posts || [];
            if (posts.length == 0) {
                $scope.message = "the end of the list...";
                $timeout(function() {
                    $scope.message = "";
                }, 2000);
                return;
            } else {
                $location.url("/" + q + "/" + posts[0].key);
            }
        });
    };
    $scope.$watch("key", function(key) {
        if (!key) {
            $scope.loading = false;
            return;
        }
        bbs.get(key).success(function(d) {
            $scope.message = d.error;
            $scope.post = d.result || {};
            var parsed = bbs.parse($scope.post.content)
            $scope.post.header = parsed.header;
            $scope.post.body = parsed.body;
            $scope.loading = false;
        });
    });
})
.controller("Put", function($scope, $routeParams, $location, bbs) {
    $scope.error = "";
    $scope.key = $routeParams.key || "";
    $scope.rev = 0;
    if ($scope.key) {
        $scope.content = "Loading...";
        bbs.get($scope.key).success(function(d) {
            $scope.error = d.error;
            if (d.result) {
                $scope.rev = d.result.rev;
                $scope.content = d.result.content;
            }
        })
    } else {
        var tags = $routeParams.tags || [];
        if (tags.join("").length == 0) tags = ["test"];
        var header = {
            title: $routeParams.title || "Title",
            authors: [$scope.user],
            tags: tags
        };
        if ($routeParams.reply) header.reply = $routeParams.reply;
        if ($routeParams.thread) header.thread = $routeParams.thread;
        $scope.content = "---\n" + jsyaml.safeDump(header, {flowLevel: 1});
        $scope.content += "---\nMarkdown Content";
    }

    $scope.submit = function() {
        bbs.put($scope.key, $scope.rev+1, $scope.content).success(function(d) {
            if (d.error) {
                $scope.error = d.error;
            } else if (d.result) {
                var h = bbs.parse($scope.content).header;
                var q = h && h.thread || d.result;
                $location.url("/" + q + "/" + d.result);
            }
        })
    };
})
.directive("post", function ($sce) {
    var markdown = new Showdown.converter({ extensions: ['tagbbs'] });

    return {
        require: "ngModel",
        restrict: "E",
        templateUrl: "post.html",
        link: function (scope, element, attrs, ngModel) {
            ngModel.$render = function() {
                var post = ngModel.$viewValue;
                scope.rev = post.rev;
                scope.timestamp = post.timestamp;
                scope.title = "";
                scope.tags = [];
                scope.authors = [];
                scope.body = "";
                scope.source = post.content;

                if (post.header) {
                    scope.title = post.header.title;
                    scope.tags = post.header.tags;
                    scope.authors = post.header.authors;
                    if (post.header.raw) scope.source = post.body;
                    else scope.body = $sce.trustAsHtml(markdown.makeHtml(post.body));
                }
            };
        }
    };
})
.directive("codemirror", function(isMobile) {
return {
        require: "ngModel",
        restrict: "A",
        link: function (scope, elm, attrs, ngModel) {
            // Disable codemirror for mobile. The touch interface does not seem play well with it.
            if (isMobile.any()) return;
            var codemirror = CodeMirror.fromTextArea(elm[0])
            codemirror.on("change", function(mirror) {
                var newValue = mirror.getValue();
                if (newValue !== ngModel.$viewValue) {
                    ngModel.$setViewValue(newValue);
                    scope.$apply();
                }
            });
            ngModel.$render = function() {
                codemirror.setValue(ngModel.$viewValue);
            };
       }
    };
})
.factory("bbs", function($http, serviceEndpoint) {
    var sid = "";
    var api = function(name, data) {
        data = data || {};
        data.session = sid;
        var promise = $http({
            method: 'POST',
            url: serviceEndpoint + "/" + name,
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            transformRequest: function(obj) {
                var str = [];
                for(var p in obj)
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
                return str.join("&");
            },
            data: data,
            timeout: 20000
        });
        return promise;
    };
    return {
        login: function(user, pass) {
            return api("login", {user: user, pass: pass}).success(function(d) {
                if (d.result) {
                    sid = d.result;
                }
                return d;
            });
        },
        logout: function() {
            return api("logout");
        },
        version: function() {
            return api("version");
        },
        who: function() {
            return api("who");
        },
        register: function() {
            return api("register");
        },
        passwd: function(pass) {
            return api("passwd", {pass: pass});
        },
        list: function(query) {
            return api("list", {query: query});
        },
        get: function(key) {
            return api("get", {key: key});
        },
        put: function(key, rev, content) {
            return api("put", {key:key, rev: rev, content: content});
        },
        session: function(_sid) {
            oldsid = sid;
            if (typeof _sid != 'undefined') {
                sid = _sid
            }
            return oldsid;
        },
        parse: function(source) {
            source = source || "";
            if (source.length == 0) return {header:null, body:""};
            var trimmed = source.trimLeft();
            if (trimmed.substring(0, 4) == "---\n") {
                var headerEnd = trimmed.indexOf("\n---\n");
                if (headerEnd > 0) {
                    var header = trimmed.substring(0, headerEnd);
                    var body = trimmed.substring(headerEnd + 5);
                    try {
                        return {header: jsyaml.safeLoad(header), body: body}
                    } catch (e) {
                        console.log(e);
                    }
                }
            }
            return {header: null, body: source}
        }
    };
})
.config(function($httpProvider) {
    $httpProvider.interceptors.push(function($q, $timeout, ngProgressLite) {
        var active = 0;
        var start = function() {
            active++;
            $timeout(function() {
                if (active > 0) {
                    ngProgressLite.start();
                    ngProgressLite.inc();
                }
            }, 100);
        };
        var finish = function() {
            active--;
            $timeout(function() {
                if (active == 0) {
                    ngProgressLite.done();
                }
            }, 100);
        };
        return {
            'request': function(config) {
              start()
              return config || $q.when(config);
            },
            'response': function(response) {
              finish();
              return response || $q.when(response);
            },
           'responseError': function(rejection) {
              finish();
              return $q.reject(rejection);
            }
        };
    });
})
.value("serviceEndpoint", "https://secure.thinxer.com:8023")
.factory("isMobile", function() {
    var isMobile = {
        Android: function() {
            return navigator.userAgent.match(/Android/i);
        },
        BlackBerry: function() {
            return navigator.userAgent.match(/BlackBerry/i);
        },
        iOS: function() {
            return navigator.userAgent.match(/iPhone|iPad|iPod/i);
        },
        Opera: function() {
            return navigator.userAgent.match(/Opera Mini/i);
        },
        Windows: function() {
            return navigator.userAgent.match(/IEMobile/i);
        },
        any: function() {
            return (isMobile.Android() || isMobile.BlackBerry() || isMobile.iOS() || isMobile.Opera() || isMobile.Windows());
        }
    };
    return isMobile;
})
;

//  Showdown TagBBS Extension
//  @username   ->  <a href="#/@/user:username">@username</a>
//  #hashtag    ->  <a href="#/hashtag">#hashtag</a>
(function(){
    var tagbbs = function(converter) {
        return [

            // #hashtag syntax, must be prefixed with a whitespace, to prevent collisions with #titles
            { type: 'lang', regex: '([ \\t])#([^\\s]+)', replace: function(match, whitespace, tag) {
                return whitespace + '<a href="#/' + tag.toLowerCase() + '">#' + tag + '</a>';
            }},

            // @username syntax, can be placed anywhere
            { type: 'lang', regex: '(\\s)@([^\\s]+)', replace: function(match, whitespace, username) {
                return whitespace + '<a href="#/@/user:' + username + '">@' + username + '</a>';
            }},

            // Escaped @'s
            { type: 'lang', regex: '\\\\@', replace: '@' }
        ];
    };

    // Client-side export
    if (typeof window !== 'undefined' && window.Showdown && window.Showdown.extensions) { window.Showdown.extensions.tagbbs = tagbbs; }
}());
