(defcolumns X Y ST)
(defun (getX) X)
(defun (getY) Y)
(defun (nextX) (shift X 1))

(defconstraint c1 () (* ST (+ (shift (getX) 1) Y)))
(defconstraint c2 () (* ST (+ (shift X 1) (getY))))
(defconstraint c3 () (* ST (+ (shift (getX) 1) (getY))))
(defconstraint c4 () (* ST (+ (nextX) Y)))
