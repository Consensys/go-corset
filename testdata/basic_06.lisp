(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns (X :i16))
(defconstraint c1 () (vanishes! (* X (- X 1))))
