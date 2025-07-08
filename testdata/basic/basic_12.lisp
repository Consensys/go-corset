(defcolumns (X_LO :i8) (X_HI :i8) (Y :i16))
;;
(defconstraint c1 () (== Y X_HI::X_LO))
