# Preserve two-level comment threads

The blog will support replies within comment threads instead of a purely flat comment list, but replies are limited to two visual levels: top-level comments plus replies. This keeps reader interaction expressive for technical discussions while avoiding recursive thread rendering; deeper reply intent is represented with `reply_to_user_id` and displayed as an `@user` tag, with unlimited nesting marked as a future optimization.
