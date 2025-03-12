;;error:6:21-25:qualified access not permitted here
;;error:6:21-25:expected loobean constraint (found u16)
(module m1)
(defcolumns (X :i16))

(defconstraint c () m1.X)
