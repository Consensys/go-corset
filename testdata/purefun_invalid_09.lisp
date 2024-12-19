;;error:6:39-40:expected type u1 (found 𝔽)
;;error:7:39-40:expected type u1 (found u16)
(defpurefun ((eq :binary@loob :force) (x :binary) (y :binary)) (^ (- x y) 2))
;;
(defcolumns (X :binary@loob) Y (Z :i16))
(defconstraint c1 () (* (- X 1) (eq X Y)))
(defconstraint c2 () (* (- X 1) (eq X Z)))
