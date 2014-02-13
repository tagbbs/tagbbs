TagBBS
======

TagBBS is a tag-based discussion system, based on discussions with @HenryHu.

Design
------

Aiming to be simple and extensible, the system works against a simple key-value store interface, and should have no internal states. The only requirement of the key-value store is to support a revision field for each key, and update the key only if the new revision is incremented by one.

This package is the core logic of the system and based on it various external interfaces can be built.

A list of interfaces:

+ `apibbsd` provides a CORS enabled web API endpoint. `webui` is an web app backed by this interface.
+ `sshbbsd` provides a simple SSH service.

Key Types
---------

User Editable Keys:

+ `post:id` are the posts. Readable by
+ `user:*` are user profiles of each user.

System Keys:

+ `bbs:*` are the meta data for the BBS.
    - `bbs:name` is the name of the BBS.
    - `bbs:nextid` is the next available post id.
    - `bbs:users` is a list of all available users.
+ `userpass:*` are the user password hashes.

Index Keys:

+ `tag:*` are the post lists of every tags.


Post
----

User Editable Keys are mapped to Posts. A Post is a user editable file, usually consists of a YAML header and a Markdown body. A sample Post:

```
---
title: Hello World
tags: [test]
authors: [thinxer]
---

There goes the content of the post.
```

The content of the post should be encoded in UTF-8 and use Unix line-ending.

The YAML header is not limited to the predefined keys. It's okay to abuse it for extra information. However, those predefined keys have special meanings and will be respected by the system.

Here is a list of predefined keys:

+ `title` is the title of the the post.
+ `tags` is a list of tags that this post should be listed under.
+ `authors` is a list of users that this post can be edited with. It is not required to include yourself in this list. As a consequence, you cannot edit this post in the future.

Besides the `Content`, a Post also has a `Rev` field, which must be incremented by 1 on each edit (an empty post always has revision 0). Upon each edit, a `Timestamp` will be marked on the post.
