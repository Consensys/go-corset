(defpurefun ((vanishes! :@loob) x) x)

(defpurefun ((or! :@loob) x y) (* x y))
(defpurefun ((eq! :@loob) x y) (- x y))

(defpurefun (prev x) (shift x -1))
(defpurefun (next x) (shift x 1))

;;
(defpurefun (if-not-eq lhs rhs then)
    (if
     (eq! lhs rhs)
     then))
