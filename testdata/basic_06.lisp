(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns X)
(defconstraint c1 () (vanishes! (* X (- X 1))))
