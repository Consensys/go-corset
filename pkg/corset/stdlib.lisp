(defpurefun ((vanishes! :@loob) x) x)

(defpurefun ((or! :@loob) x y) (* x y))
(defpurefun ((eq! :@loob) x y) (- x y))

(defpurefun (prev x) (shift x -1))
(defpurefun (next x) (shift x 1))

;;
(defpurefun (if-not-eq lhs rhs then)
    (if
     (eq! lhs rhs)
     ;; True branch
     0
     ;; False branch
     then))

(defpurefun (if-eq lhs rhs then)
    (if
     (eq! lhs rhs)
     ;; True branch
     then))

(defpurefun (if-eq-else lhs rhs then else)
    (if
     (eq! lhs rhs)
     ;; True branch
     then
     ;; False branch
     else))
