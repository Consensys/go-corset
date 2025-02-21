;;error:4:16-21:interleaved column access not permitted
(defcolumns X Y)
(definterleaved Z (X Y))
(defsorted s1 ((↓ Z)))
