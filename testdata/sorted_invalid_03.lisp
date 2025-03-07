;;error:4:16-21:interleaved column access not permitted
(defcolumns (X :i16) (Y :i16))
(definterleaved Z (X Y))
(defsorted s1 ((↓ Z)))
