;;error:5:30-35:incompatible length multiplier
;;
(defcolumns (X :i16) (Y :i16))
(definterleaved Z (X Y))
(defpermutation (A B) ((+ X) (+ Z)))
