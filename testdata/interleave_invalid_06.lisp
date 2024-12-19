;;error:5:27-28:conflicting context
;;error:5:22-29:expected loobean constraint (found 𝔽)
(defcolumns X Y)
(definterleaved A (X Y))
(defconstraint c1 () (+ A X))
