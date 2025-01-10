;;error:6:21-25:qualified access not permitted here
;;error:6:21-25:expected loobean constraint (found 𝔽)
(module m1)
(defcolumns X)

(defconstraint c () m1.X)
