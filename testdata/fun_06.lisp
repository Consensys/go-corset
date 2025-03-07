(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns X Y ST)
(defun (getX) X)
(defun (getY) Y)
(defun (nextX) (shift X 1))

(defconstraint c1 () (vanishes! (* ST (+ (shift (getX) 1) Y))))
(defconstraint c2 () (vanishes! (* ST (+ (shift X 1) (getY)))))
(defconstraint c3 () (vanishes! (* ST (+ (shift (getX) 1) (getY)))))
(defconstraint c4 () (vanishes! (* ST (+ (nextX) Y))))
