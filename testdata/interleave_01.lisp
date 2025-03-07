(defcolumns (X :i16@loob) (Y :i16@loob))
(definterleaved Z (X Y))
(defconstraint c1 () Z)
