(defcolumns (X :@loob) (Y :@loob))
(definterleaved Z (X Y))
(defconstraint c1 () Z)
