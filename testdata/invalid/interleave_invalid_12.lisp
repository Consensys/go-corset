;;error:5:20-21:invalid source column
;;
(defcolumns (Y :i16))
(defconst X 1)
(definterleaved Z (X Y))
