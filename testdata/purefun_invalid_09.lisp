(defpurefun ((eq :binary@loob :force) (x :binary) (y :binary)) (^ (- x y) 2))
;;
(defcolumns (X :binary@loob) Y (Z :i16))
(defconstraint c1 () (* (- X 1) (eq X Y)))
(defconstraint c2 () (* (- X 1) (eq X Z)))
