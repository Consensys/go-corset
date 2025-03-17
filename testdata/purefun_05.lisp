(defpurefun ((eq :binary) (x :binary) (y :binary)) (^ (- x y) 2))
;;
(defcolumns (X :binary) (Y :binary))
;; X == 1 || X == Y
(defconstraint c1 () (* (- X 1) (eq X Y)))
