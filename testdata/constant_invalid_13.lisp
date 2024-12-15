;;error:5:41-44:not permitted in pure context
;;error:5:71-74:not permitted in pure context
;;error:5:22-77:expected loobean constraint (found 𝔽)
(defcolumns CT ONE)
(defconstraint c1 () (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE))))
