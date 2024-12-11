(defpurefun ((vanishes! :@loob) x) x)
;;
(defcolumns X)
(defconstraint c1 () (vanishes! (* X (- X 1))))
