;;error:2:1-2:blah
(defcolumns X Y)
(definterleaved Z (X Y))
(defpermutation (A B) ((+ Z) (+ Y)))
