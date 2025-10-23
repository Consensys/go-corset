(defcolumns (X :i16) (Y :i16) (Z :i16) (R :i16))
;; R::Z == X * Y
(defconstraint c1 () (== (:: R Z) (* X Y)))
