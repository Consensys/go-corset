;;error:6:26-30:qualified access not permitted here
;;error:6:32-33:expected loobean constraint (found 𝔽)
(module m1)
(defcolumns X)

(defconstraint c (:guard m1.X) X)
