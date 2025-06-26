(defcolumns (X_LO :i128) (X_HI :i128) (Y :i256))
;;
(defconstraint c1 () (== Y X_HI::X_LO))
