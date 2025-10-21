;;
(defcolumns (X :i16) (Y :i16) (ST :i16))
(defun (getX) X)
(defun (getY) Y)
(defun (nextX) (shift X 1))

(defconstraint c1 () (== 0 (* ST (- (shift (getX) 1) Y))))
(defconstraint c2 () (== 0 (* ST (- (shift X 1) (getY)))))
(defconstraint c3 () (== 0 (* ST (- (shift (getX) 1) (getY)))))
(defconstraint c4 () (== 0 (* ST (- (nextX) Y))))
