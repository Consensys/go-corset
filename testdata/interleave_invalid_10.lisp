;;error:5:24-29:fixed-width type required
;;error:5:30-35:fixed-width type required
(defcolumns X Y)
(definterleaved Z (X Y))
(defpermutation (A B) ((+ X) (+ Z)))
